package repository

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"gorm.io/gorm"
)

type commentModel struct {
	gorm.Model
	PostID  uint   `gorm:"notNull"`
	UserID  uint   `gorm:"notNull"`
	Content string `gorm:"type:text;notNull"`
}

func (m *commentModel) TableName() string {
	return "comments"
}

func (m *commentModel) fromDomain(c *domain.Comment) {
	m.ID = uint(c.ID)
	m.PostID = uint(c.PostID)
	m.UserID = uint(c.UserID)
	m.Content = c.Content
}

type commentRepository struct {
	db *gorm.DB
}

// NewCommentRepository creates a new instance of CommentRepository.
func NewCommentRepository(db *gorm.DB) ports.CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	var m commentModel
	m.fromDomain(comment)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	comment.ID = int(m.ID)
	comment.CreatedAt = m.CreatedAt
	comment.UpdatedAt = m.UpdatedAt
	return nil
}
