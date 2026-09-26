package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/billykore/project-one/internal/publishing/domain"
	"github.com/billykore/project-one/internal/publishing/ports"
	"gorm.io/gorm"
)

type commentModel struct {
	gorm.Model
	PostID uint64 `gorm:"notNull"`
	UserID int    `gorm:"notNull"`
	// Username is populated by read queries joining users; it is never persisted here.
	Username string `gorm:"-"`
	Content  string `gorm:"type:text;notNull"`
}

func (m *commentModel) TableName() string {
	return "comments"
}

func (m *commentModel) fromDomain(c *domain.Comment) {
	m.ID = uint(c.ID)
	m.PostID = uint64(c.PostID)
	m.UserID = c.UserID
	m.Content = c.Content
}

func (m *commentModel) toDomain() *domain.Comment {
	return &domain.Comment{
		ID:        int(m.ID),
		PostID:    int(m.PostID),
		UserID:    m.UserID,
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
	username := comment.Username
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return fmt.Errorf("%w: %v", problem.ErrRepositoryFailure, err)
	}
	*comment = *m.toDomain()
	comment.Username = username
	return nil
}

func (r *commentRepository) GetByPostID(ctx context.Context, postID int) ([]*domain.Comment, error) {
	var models []commentModel
	err := r.db.WithContext(ctx).Table("comments").
		Select("comments.*, users.username").
		Joins("INNER JOIN users ON users.id = comments.user_id").
		Where("comments.post_id = ? AND comments.deleted_at IS NULL", postID).
		Order("comments.created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("%w: %v", problem.ErrRepositoryFailure, err)
	}

	comments := make([]*domain.Comment, 0, len(models))
	for _, m := range models {
		comments = append(comments, m.toDomain())
	}
	return comments, nil
}

func (r *commentRepository) GetByID(ctx context.Context, id int) (*domain.Comment, error) {
	var m commentModel
	err := r.db.WithContext(ctx).Table("comments").
		Select("comments.*, users.username").
		Joins("INNER JOIN users ON users.id = comments.user_id").
		Where("comments.id = ? AND comments.deleted_at IS NULL", id).
		Take(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, problem.ErrCommentNotFound
		}
		return nil, fmt.Errorf("%w: %v", problem.ErrRepositoryFailure, err)
	}
	return m.toDomain(), nil
}

func (r *commentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	var m commentModel
	m.fromDomain(comment)
	m.ID = uint(comment.ID)

	// Select only content to avoid side updates (e.g. author changes)
	err := r.db.WithContext(ctx).Model(&m).
		Select("Content").
		Updates(&m).Error
	if err != nil {
		return fmt.Errorf("%w: %v", problem.ErrRepositoryFailure, err)
	}

	comment.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *commentRepository) Delete(ctx context.Context, id int) error {
	err := r.db.WithContext(ctx).Delete(&commentModel{}, id).Error
	if err != nil {
		return fmt.Errorf("%w: %v", problem.ErrRepositoryFailure, err)
	}
	return nil
}
