package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
)

type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) error
}

type PostService interface {
	CreatePost(ctx context.Context, userID int, title, content string, tags []string) (*domain.Post, error)
}
