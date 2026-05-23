package repository

import (
	"context"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"gorm.io/gorm"
)

type commentModel struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	PostID    uint64 `gorm:"notNull"`
	Username  string `gorm:"size:255;notNull"`
	Content   string `gorm:"type:text;notNull"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (m *commentModel) TableName() string {
	return "comments"
}

func (m *commentModel) fromDomain(c *domain.Comment) {
	m.ID = uint64(c.ID)
	m.PostID = uint64(c.PostID)
	m.Username = c.Username
	m.Content = c.Content
}

func (m *commentModel) toDomain() *domain.Comment {
	return &domain.Comment{
		ID:        int64(m.ID),
		PostID:    int64(m.PostID),
		Username:  m.Username,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt.Time,
	}
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
	*comment = *m.toDomain()
	return nil
}
