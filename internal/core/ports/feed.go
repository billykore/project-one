package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/pkg/pagination"
)

// FeedResult holds the result of a feed query.
type FeedResult struct {
	Posts      []*domain.Post
	NextCursor *pagination.Cursor
	HasMore    bool
}

// FeedUseCase is a driving port for feed-related application logic.
type FeedUseCase interface {
	GetFeed(ctx context.Context, username string, cursor *pagination.Cursor, limit int) (*FeedResult, error)
}
