package repository

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type postModel struct {
	gorm.Model
	UserID  uint           `gorm:"notNull"`
	Title   string         `gorm:"size:255;notNull"`
	Content string         `gorm:"type:text;notNull"`
	Tags    pq.StringArray `gorm:"type:text[]"`
}

func (m *postModel) TableName() string {
	return "posts"
}

func (m *postModel) fromDomain(p *domain.Post) {
	m.UserID = uint(p.UserID)
	m.Title = p.Title
	m.Content = p.Content
	m.Tags = pq.StringArray(p.Tags)
}

type postRepository struct {
	db *gorm.DB
}

// NewPostRepository creates a new instance of PostRepository.
func NewPostRepository(db *gorm.DB) ports.PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *domain.Post) error {
	var m postModel
	m.fromDomain(post)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	post.ID = int(m.ID)
	post.CreatedAt = m.CreatedAt
	post.UpdatedAt = m.UpdatedAt
	return nil
}
