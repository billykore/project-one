package repository

import (
	"context"
	"errors"

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

func (m *postModel) toDomain() *domain.Post {
	return &domain.Post{
		ID:        int(m.ID),
		UserID:    int(m.UserID),
		Title:     m.Title,
		Content:   m.Content,
		Tags:      []string(m.Tags),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt.Time,
	}
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

func (r *postRepository) GetByID(ctx context.Context, id int) (*domain.Post, error) {
	var m postModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPostNotFound
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *postRepository) GetPostsByUserID(ctx context.Context, userID int) ([]*domain.Post, error) {
	var models []postModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, err
	}

	posts := make([]*domain.Post, 0, len(models))
	for _, m := range models {
		posts = append(posts, m.toDomain())
	}
	return posts, nil
}

func (r *postRepository) Update(ctx context.Context, post *domain.Post) error {
	var m postModel
	m.ID = uint(post.ID)
	m.fromDomain(post)
	if err := r.db.WithContext(ctx).Model(&m).Select("Title", "Content", "Tags").Updates(m).Error; err != nil {
		return err
	}
	post.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *postRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&postModel{}, id).Error; err != nil {
		return err
	}
	return nil
}
