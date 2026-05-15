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
	FollowerID       int       `gorm:"primaryKey"`
	FollowerUsername string    `gorm:"primaryKey"`
	FollowedID       int       `gorm:"primaryKey"`
	FollowedUsername string    `gorm:"primaryKey"`
	CreatedAt        time.Time `gorm:"default:CURRENT_TIMESTAMP"`
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
		FollowerUsername: follow.FollowerUsername,
		FollowedUsername: follow.FollowedUsername,
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

func (r *followRepository) GetFollowing(ctx context.Context, followerUsername string, limit, offset int) ([]domain.Following, error) {
	var results []domain.Following
	err := r.db.WithContext(ctx).Table("follows").
		Select("users.username, users.first_name, users.last_name, follows.created_at AS followed_at, (mutual.follower_username IS NOT NULL) AS is_mutual").
		Joins("INNER JOIN users ON users.username = follows.followed_username").
		Joins("LEFT JOIN follows AS mutual ON mutual.follower_username = follows.followed_username AND mutual.followed_username = follows.follower_username").
		Where("follows.follower_username = ?", followerUsername).
		Order("follows.created_at DESC, follows.followed_username DESC").
		Limit(limit).Offset(offset).
		Scan(&results).Error

	return results, err
}

func (r *followRepository) GetFollowers(ctx context.Context, followedUsername string, limit, offset int) ([]domain.Follower, error) {
	var results []domain.Follower
	err := r.db.WithContext(ctx).Table("follows").
		Select("users.username, users.first_name, users.last_name, follows.created_at AS followed_at, (mutual.follower_username IS NOT NULL) AS is_mutual").
		Joins("INNER JOIN users ON users.username = follows.follower_username").
		Joins("LEFT JOIN follows AS mutual ON mutual.follower_username = follows.follower_username AND mutual.followed_username = follows.followed_username").
		Where("follows.followed_username = ?", followedUsername).
		Order("follows.created_at DESC, follows.follower_username DESC").
		Limit(limit).Offset(offset).
		Scan(&results).Error

	return results, err
}

func (r *followRepository) Delete(ctx context.Context, followerUsername, followedUsername string) error {
	result := r.db.WithContext(ctx).
		Where("follower_username = ? AND followed_username = ?", followerUsername, followedUsername).
		Delete(&followModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFollowing
	}
	return nil
}
