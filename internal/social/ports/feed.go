package ports

import (
	"context"

	vo "github.com/billykore/project-one/internal/platform/pagination"
	publishingdomain "github.com/billykore/project-one/internal/publishing/domain"
)

// FeedResult holds the result of a feed query.
type FeedResult struct {
	Posts      []*publishingdomain.Post
	NextCursor *vo.Cursor
	HasMore    bool
}

// FeedUseCase is a driving port for feed-related application logic.
type FeedUseCase interface {
	// GetFeed retrieves a paginated list of posts for the authenticated user,
	// including posts from users they follow and their own posts.
	GetFeed(ctx context.Context, userID int, cursor *vo.Cursor, limit int) (*FeedResult, error)
}
