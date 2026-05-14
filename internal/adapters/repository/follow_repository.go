package repository

import (
	"context"
	"errors"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"gorm.io/gorm"
)

type followModel struct {
	FollowerID int       `gorm:"primaryKey"`
	FollowedID int       `gorm:"primaryKey"`
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
			return domain.ErrAlreadyFollowing
		}
		return domain.ErrUserNotFound
	}
	follow.CreatedAt = m.CreatedAt
	return nil
}

func (r *followRepository) IsFollowing(ctx context.Context, followerID, followedID int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&followModel{}).
		Where("follower_id = ? AND followed_id = ?", followerID, followedID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *followRepository) GetFollowing(ctx context.Context, followerID int, limit, offset int) ([]domain.Following, error) {
	var results []domain.Following
	err := r.db.WithContext(ctx).Table("follows").
		Select("users.id, users.first_name, users.last_name, follows.created_at AS followed_at, (mutual.follower_id IS NOT NULL) AS is_mutual").
		Joins("INNER JOIN users ON users.id = follows.followed_id").
		Joins("LEFT JOIN follows AS mutual ON mutual.follower_id = follows.followed_id AND mutual.followed_id = follows.follower_id").
		Where("follows.follower_id = ?", followerID).
		Order("follows.created_at DESC").
		Limit(limit).Offset(offset).
		Scan(&results).Error

	return results, err
}

func (r *followRepository) GetFollowers(ctx context.Context, followedID int, limit, offset int) ([]domain.Follower, error) {
	var results []domain.Follower
	err := r.db.WithContext(ctx).Table("follows").
		Select("users.id, users.first_name, users.last_name, follows.created_at AS followed_at, (mutual.follower_id IS NOT NULL) AS is_mutual").
		Joins("INNER JOIN users ON users.id = follows.follower_id").
		Joins("LEFT JOIN follows AS mutual ON mutual.follower_id = follows.followed_id AND mutual.followed_id = follows.follower_id").
		Where("follows.followed_id = ?", followedID).
		Order("follows.created_at DESC").
		Limit(limit).Offset(offset).
		Scan(&results).Error

	return results, err
}
