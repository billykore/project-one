package repository

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"gorm.io/gorm"
)

type commentModel struct {
	gorm.Model
	PostID   uint64 `gorm:"notNull"`
	Username string `gorm:"size:255;notNull"`
	Content  string `gorm:"type:text;notNull"`
}

func (m *commentModel) TableName() string {
	return "comments"
}

func (m *commentModel) fromDomain(c *domain.Comment) {
	m.ID = uint(c.ID)
	m.PostID = uint64(c.PostID)
	m.Username = c.Username
	m.Content = c.Content
}

func (m *commentModel) toDomain() *domain.Comment {
	return &domain.Comment{
		ID:        int(m.ID),
		PostID:    int(m.PostID),
		Username:  m.Username,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
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

func (r *commentRepository) GetByPostID(ctx context.Context, postID int) ([]*domain.Comment, error) {
	var models []commentModel
	err := r.db.WithContext(ctx).
		Where("post_id = ?", postID).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	comments := make([]*domain.Comment, 0, len(models))
	for _, m := range models {
		comments = append(comments, m.toDomain())
	}
	return comments, nil
}
