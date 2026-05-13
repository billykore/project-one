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
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return domain.ErrUserNotFound
		}
		return err
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
