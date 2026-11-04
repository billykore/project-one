package ports

import (
	"context"

	vo "github.com/billykore/project-one/internal/pkg/pagination"
	"github.com/billykore/project-one/internal/social/domain"
)

// AccountLookup is social's narrow identity lookup contract.
type AccountLookup interface {
	GetUserByID(ctx context.Context, id int) (*domain.Account, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.Account, error)
}

// TimelineReader is social's read-only view of published posts.
type TimelineReader interface {
	// GetFeed returns posts authored by userIDs, newest first, after cursor.
	GetFeed(ctx context.Context, userIDs []int, cursor *vo.Cursor, limit int) ([]*domain.TimelinePost, error)
}
