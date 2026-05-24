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
	Username string         `gorm:"size:255;notNull"`
	Title    string         `gorm:"size:255;notNull"`
	Content  string         `gorm:"type:text;notNull"`
	Tags     pq.StringArray `gorm:"type:text[]"`
}

func (m *postModel) TableName() string {
	return "posts"
}

func (m *postModel) fromDomain(p *domain.Post) {
	m.Username = p.Username
	m.Title = p.Title
	m.Content = p.Content
	m.Tags = pq.StringArray(p.Tags)
}

func (m *postModel) toDomain() *domain.Post {
	return &domain.Post{
		ID:        int(m.ID),
		Username:  m.Username,
		Title:     m.Title,
		Content:   m.Content,
		Tags:      []string(m.Tags),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
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

func (r *postRepository) GetByID(ctx context.Context, username string, id int) (*domain.Post, error) {
	var m postModel
	err := r.db.WithContext(ctx).
		Where("username = ? AND id = ?", username, id).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPostNotFound
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *postRepository) GetByIDOnly(ctx context.Context, id int) (*domain.Post, error) {
	var m postModel
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPostNotFound
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *postRepository) GetUserPosts(ctx context.Context, username string, limit, offset int) ([]*domain.Post, error) {
	var models []postModel
	query := r.db.WithContext(ctx).Where("username = ?", username)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	posts := make([]*domain.Post, 0, len(models))
	for _, m := range models {
		posts = append(posts, m.toDomain())
	}
	return posts, nil
}

func (r *postRepository) Update(ctx context.Context, username string, post *domain.Post) error {
	var m postModel
	m.ID = uint(post.ID)
	m.fromDomain(post)
	if err := r.db.WithContext(ctx).Model(&m).
		Select("Title", "Content", "Tags").
		Where("username = ? AND id = ?", username, post.ID).
		Updates(&m).Error; err != nil {
		return err
	}
	post.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *postRepository) Delete(ctx context.Context, username string, id int) error {
	if err := r.db.WithContext(ctx).Where("username = ? AND id = ?", username, id).Delete(&postModel{}).Error; err != nil {
		return err
	}
	return nil
}
