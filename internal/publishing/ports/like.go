package ports

import (
	"context"

	"github.com/billykore/project-one/internal/publishing/domain"
)

// LikeRepository is a driven port for post like persistence.
type LikeRepository interface {
	// SetLiked atomically creates or removes a like and updates the post's
	// denormalized count in the same transaction. changed is false for an
	// idempotent request; likeCount is the persisted count after the operation.
	SetLiked(ctx context.Context, postID int, userID int, liked bool) (int, bool, error)
	// Create adds a like to the database. It should return an error if the like already exists.
	Create(ctx context.Context, like *domain.Like) error
	// Delete removes a like from the database. It should return an error if the like does not exist.
	Delete(ctx context.Context, postID int, userID int) error
	// Exists checks if a like exists for the given post and user IDs.
	// It should return true if the like exists, false otherwise.
	Exists(ctx context.Context, postID int, userID int) (bool, error)
	// CountByPostID returns the total number of likes for a given post ID.
	CountByPostID(ctx context.Context, postID int) (int, error)
}
