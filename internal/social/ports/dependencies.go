package ports

import (
	"context"

	vo "github.com/billykore/project-one/internal/pkg/pagination"
	publishingdomain "github.com/billykore/project-one/internal/publishing/domain"
	"github.com/billykore/project-one/internal/social/domain"
)

// AccountLookup is social's narrow identity lookup contract.
type AccountLookup interface {
	GetUserByID(ctx context.Context, id int) (*domain.Account, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.Account, error)
}

// TimelineReader is social's read-only view of published posts.
type TimelineReader interface {
	GetFeed(ctx context.Context, userIDs []int, cursor *vo.Cursor, limit int) ([]*publishingdomain.Post, error)
}
