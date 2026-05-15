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
	// GetFollowing fetches the paginated list of users being followed by a specific user.
	GetFollowing(ctx context.Context, followerID int, limit, offset int) ([]domain.Following, error)
	// GetFollowers fetches the paginated list of users following a specific user.
	GetFollowers(ctx context.Context, followedID int, limit, offset int) ([]domain.Follower, error)
	// Delete removes an existing follow relationship.
	Delete(ctx context.Context, followerID, followedID int) error
}

// FollowUseCase defines the interface for follow-related business logic.
type FollowUseCase interface {
	// Follow handles the logic for a user following another user.
	Follow(ctx context.Context, followerUsername, followedUsername string) (*domain.Follow, error)
	// GetFollowing handles the logic for getting the following list of a user.
	GetFollowing(ctx context.Context, followerUsername string, limit, offset int) ([]domain.Following, error)
	// GetFollowers handles the logic for getting the followers list of a user.
	GetFollowers(ctx context.Context, followedUsername string, limit, offset int) ([]domain.Follower, error)
	// Unfollow handles the logic for a user unfollowing another user.
	Unfollow(ctx context.Context, followerUsername, followedUsername string) error
}
