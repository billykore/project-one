package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
)

// LikeRepository is a driven port for post like persistence.
type LikeRepository interface {
	// Create adds a like to the database. It should return an error if the like already exists.
	Create(ctx context.Context, like *domain.Like) error
	// Delete removes a like from the database. It should return an error if the like does not exist.
	Delete(ctx context.Context, postID int, username string) error
	// Exists checks if a like exists for the given post ID and username.
	// It should return true if the like exists, false otherwise.
	Exists(ctx context.Context, postID int, username string) (bool, error)
	// CountByPostID returns the total number of likes for a given post ID.
	CountByPostID(ctx context.Context, postID int) (int, error)
}
