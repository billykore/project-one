package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
)

// CommentRepository is a driven port for comment persistence.
type CommentRepository interface {
	// Create saves a new comment to the repository.
	Create(ctx context.Context, comment *domain.Comment) error
	// GetByPostID retrieves all comments for a specific post.
	GetByPostID(ctx context.Context, postID int) ([]*domain.Comment, error)
	// GetByID retrieves a comment by its ID.
	GetByID(ctx context.Context, id int) (*domain.Comment, error)
	// Update updates a comment's details.
	Update(ctx context.Context, comment *domain.Comment) error
	// Delete removes a comment by its ID from the repository.
	Delete(ctx context.Context, id int) error
}

// CommentUseCase is a driving port for comment-related application logic.
type CommentUseCase interface {
	// AddComment creates a new comment on a post.
	AddComment(ctx context.Context, postID int, username string, content string) error
	// GetCommentsByPostID retrieves all comments for a specific post.
	GetCommentsByPostID(ctx context.Context, postID int) ([]*domain.Comment, error)
	// EditComment updates an existing comment's content.
	EditComment(ctx context.Context, id int, username string, content string) error
	// DeleteComment deletes a comment if it belongs to the user.
	DeleteComment(ctx context.Context, id int, username string) error
}
