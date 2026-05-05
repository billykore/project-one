package repository

import (
	"context"
	"github.com/billykore/project-one/internal/app/post/core/domain"
	"github.com/billykore/project-one/internal/app/post/core/ports"
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

type postgresPostRepository struct {
	db *gorm.DB
}

func NewPostgresPostRepository(db *gorm.DB) ports.PostRepository {
	return &postgresPostRepository{db: db}
}

func (r *postgresPostRepository) Create(ctx context.Context, post *domain.Post) error {
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
