package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
)

// FollowRepository defines the interface for follow-related data access.
type FollowRepository interface {
	// Create persists a new follow relationship.
	Create(ctx context.Context, follow *domain.Follow) error
	// IsFollowing checks if a follower-followed relationship already exists.
	IsFollowing(ctx context.Context, followerID, followedID int) (bool, error)
}

// FollowUseCase defines the interface for follow-related business logic.
type FollowUseCase interface {
	// Follow handles the logic for a user following another user.
	Follow(ctx context.Context, followerID, followedID int) (*domain.Follow, error)
}
