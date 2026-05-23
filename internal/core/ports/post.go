package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
)

// PostRepository is a driven port for post persistence.
type PostRepository interface {
	// Create saves a new post to the repository.
	Create(ctx context.Context, post *domain.Post) error
	// GetByID retrieves a post by its ID.
	GetByID(ctx context.Context, username string, id int) (*domain.Post, error)
	// GetByIDOnly retrieves a post by its ID without checking owner.
	GetByIDOnly(ctx context.Context, id int) (*domain.Post, error)
	// GetUserPosts retrieves all posts for a specific user.
	GetUserPosts(ctx context.Context, username string, limit, offset int) ([]*domain.Post, error)
	// Update updates an existing post in the repository.
	Update(ctx context.Context, username string, post *domain.Post) error
	// Delete removes a post from the repository.
	Delete(ctx context.Context, username string, id int) error
}

// PostUseCase is a driving port for post-related application logic.
type PostUseCase interface {
	// CreatePost creates a new post with the given details.
	CreatePost(ctx context.Context, username string, title, content string, tags []string) (*domain.Post, error)
	// GetPostByID retrieves a post by its ID for a specific user.
	GetPostByID(ctx context.Context, username string, postID int) (*domain.Post, error)
	// GetPosts retrieves all posts for a specific user.
	GetPosts(ctx context.Context, username string, limit, offset int) ([]*domain.Post, error)
	// UpdatePost updates an existing post for a specific user.
	UpdatePost(ctx context.Context, username string, postID int, title, content string) (*domain.Post, error)
	// DeletePost removes a post for a specific user.
	DeletePost(ctx context.Context, username string, postID int) error
}
