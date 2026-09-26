package ports

import (
	"context"

	identitydomain "github.com/billykore/project-one/internal/identity/domain"
	vo "github.com/billykore/project-one/internal/platform/pagination"
	publishingdomain "github.com/billykore/project-one/internal/publishing/domain"
)

// AccountLookup is social's narrow identity lookup contract.
type AccountLookup interface {
	GetUserByID(ctx context.Context, id int) (*identitydomain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*identitydomain.User, error)
}

// TimelineReader is social's read-only view of published posts.
type TimelineReader interface {
	GetFeed(ctx context.Context, userIDs []int, cursor *vo.Cursor, limit int) ([]*publishingdomain.Post, error)
}
