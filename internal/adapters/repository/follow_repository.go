package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	vo "github.com/billykore/project-one/internal/core/valueobject"
	"gorm.io/gorm"
)

type followModel struct {
	FollowerID int
	FollowedID int
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (m *followModel) TableName() string {
	return "follows"
}

type followRepository struct {
	db *gorm.DB
}

// NewFollowRepository creates a new instance of FollowRepository.
func NewFollowRepository(db *gorm.DB) ports.FollowRepository {
	return &followRepository{db: db}
}

func (r *followRepository) Create(ctx context.Context, follow *domain.Follow) error {
	m := followModel{
		FollowerID: follow.FollowerID,
		FollowedID: follow.FollowedID,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("%w: %v", domain.ErrAlreadyFollowing, err)
		}
		return fmt.Errorf("%w: %v", domain.ErrUserNotFound, err)
	}
	follow.CreatedAt = m.CreatedAt
	return nil
}

func (r *followRepository) GetFollowing(ctx context.Context, followerID int, cursor *vo.Cursor, limit int) ([]domain.Following, error) {
	var results []domain.Following
	query := r.db.WithContext(ctx).Table("follows").
		Select("users.username, users.first_name, users.last_name, follows.created_at AS followed_at, (mutual.follower_id IS NOT NULL) AS is_mutual").
		Joins("INNER JOIN users ON users.id = follows.followed_id").
		Joins("LEFT JOIN follows AS mutual ON mutual.follower_id = follows.followed_id AND mutual.followed_id = follows.follower_id").
		Where("follows.follower_id = ?", followerID)
	if cursor != nil && !cursor.CreatedAt.IsZero() && cursor.Key != "" {
		query = query.Where("(follows.created_at < ?) OR (follows.created_at = ? AND users.username < ?)", cursor.CreatedAt, cursor.CreatedAt, cursor.Key)
	}
	err := query.Order("follows.created_at DESC, users.username DESC").Limit(limit).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	return results, nil
}

func (r *followRepository) GetFollowers(ctx context.Context, followedID int, cursor *vo.Cursor, limit int) ([]domain.Follower, error) {
	var results []domain.Follower
	query := r.db.WithContext(ctx).Table("follows").
		Select("users.username, users.first_name, users.last_name, follows.created_at AS followed_at, (mutual.follower_id IS NOT NULL) AS is_mutual").
		Joins("INNER JOIN users ON users.id = follows.follower_id").
		Joins("LEFT JOIN follows AS mutual ON mutual.follower_id = follows.followed_id AND mutual.followed_id = follows.follower_id").
		Where("follows.followed_id = ?", followedID)
	if cursor != nil && !cursor.CreatedAt.IsZero() && cursor.Key != "" {
		query = query.Where("(follows.created_at < ?) OR (follows.created_at = ? AND users.username < ?)", cursor.CreatedAt, cursor.CreatedAt, cursor.Key)
	}
	err := query.Order("follows.created_at DESC, users.username DESC").Limit(limit).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	return results, nil
}

func (r *followRepository) Delete(ctx context.Context, followerID, followedID int) error {
	result := r.db.WithContext(ctx).
		Where("follower_id = ? AND followed_id = ?", followerID, followedID).
		Delete(&followModel{})
	if result.Error != nil {
		return fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: %v", domain.ErrNotFollowing, errors.New("no such follow relationship"))
	}
	return nil
}

func (r *followRepository) GetFollowedUserIDs(ctx context.Context, followerID int) ([]int, error) {
	var ids []int
	err := r.db.WithContext(ctx).
		Model(&followModel{}).
		Where("follower_id = ?", followerID).
		Pluck("followed_id", &ids).Error
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}
	return ids, nil
}
