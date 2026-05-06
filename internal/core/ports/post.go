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
	GetByID(ctx context.Context, id int) (*domain.Post, error)
}

// PostService is a driving port for post-related application logic.
type PostService interface {
	// CreatePost creates a new post with the given details.
	CreatePost(ctx context.Context, userID int, title, content string, tags []string) (*domain.Post, error)
	// GetPostByID retrieves a post by its ID.
	GetPostByID(ctx context.Context, id int) (*domain.Post, error)
}
