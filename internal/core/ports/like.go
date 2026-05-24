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

// LikeUseCase is a driving port for post like-related application logic.
type LikeUseCase interface {
	// ToggleLike toggles the like status for a given post ID and username.
	// It should return the new like status (liked or not) and the updated like count.
	ToggleLike(ctx context.Context, postID int, username string) (liked bool, likeCount int, err error)
	// GetLikeStatus retrieves the like status and total like count for a given post ID and username.
	GetLikeStatus(ctx context.Context, postID int, username string) (liked bool, likeCount int, err error)
}
