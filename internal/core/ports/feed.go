package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
	vo "github.com/billykore/project-one/internal/core/valueobject"
)

// FeedResult holds the result of a feed query.
type FeedResult struct {
	Posts      []*domain.Post
	NextCursor *vo.Cursor
	HasMore    bool
}

// FeedUseCase is a driving port for feed-related application logic.
type FeedUseCase interface {
	// GetFeed retrieves a paginated list of posts for the authenticated user,
	// including posts from users they follow and their own posts.
	GetFeed(ctx context.Context, username string, cursor *vo.Cursor, limit int) (*FeedResult, error)
}
