package repository

import (
	"context"
	"errors"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"gorm.io/gorm"
)

type likeModel struct {
	PostID    int       `gorm:"primaryKey"`
	Username  string    `gorm:"primaryKey;size:255"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (m *likeModel) TableName() string {
	return "post_likes"
}

func (m *likeModel) fromDomain(l *domain.Like) {
	m.PostID = l.PostID
	m.Username = l.Username
}

type likeRepository struct {
	db *gorm.DB
}

// NewLikeRepository creates a new instance of LikeRepository.
func NewLikeRepository(db *gorm.DB) ports.LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) Create(ctx context.Context, like *domain.Like) error {
	var m likeModel
	m.fromDomain(like)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrAlreadyLiked
		}
		return err
	}
	like.CreatedAt = m.CreatedAt
	return nil
}

func (r *likeRepository) Delete(ctx context.Context, postID int, username string) error {
	result := r.db.WithContext(ctx).
		Where("post_id = ? AND username = ?", postID, username).
		Delete(&likeModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotLiked
	}
	return nil
}

func (r *likeRepository) Exists(ctx context.Context, postID int, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&likeModel{}).
		Where("post_id = ? AND username = ?", postID, username).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *likeRepository) CountByPostID(ctx context.Context, postID int) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&likeModel{}).
		Where("post_id = ?", postID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
