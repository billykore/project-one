package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
	vo "github.com/billykore/project-one/internal/core/valueobject"
)

// FollowRepository defines the interface for follow-related data access.
type FollowRepository interface {
	// Create persists a new follow relationship.
	Create(ctx context.Context, follow *domain.Follow) error
	// GetFollowing fetches the cursor-paginated list of users being followed by a specific user.
	GetFollowing(ctx context.Context, followerUsername string, cursor *vo.Cursor, limit int) ([]domain.Following, error)
	// GetFollowers fetches the cursor-paginated list of users following a specific user.
	GetFollowers(ctx context.Context, followedUsername string, cursor *vo.Cursor, limit int) ([]domain.Follower, error)
	// Delete removes an existing follow relationship.
	Delete(ctx context.Context, followerUsername, followedUsername string) error
	// GetFollowedUsernames returns all usernames that the given user follows.
	GetFollowedUsernames(ctx context.Context, followerUsername string) ([]string, error)
}

// FollowingPage is a cursor-paginated following result.
type FollowingPage struct {
	Data       []domain.Following
	NextCursor *vo.Cursor
	HasMore    bool
}

// FollowersPage is a cursor-paginated follower result.
type FollowersPage struct {
	Data       []domain.Follower
	NextCursor *vo.Cursor
	HasMore    bool
}

// FollowUseCase defines the interface for follow-related business logic.
type FollowUseCase interface {
	// Follow handles the logic for a user following another user.
	Follow(ctx context.Context, followerUsername, followedUsername string) (*domain.Follow, error)
	// GetFollowing handles the logic for getting a cursor-paginated following list.
	GetFollowing(ctx context.Context, followerUsername string, cursor *vo.Cursor, limit int) (*FollowingPage, error)
	// GetFollowers handles the logic for getting a cursor-paginated followers list.
	GetFollowers(ctx context.Context, followedUsername string, cursor *vo.Cursor, limit int) (*FollowersPage, error)
	// Unfollow handles the logic for a user unfollowing another user.
	Unfollow(ctx context.Context, followerUsername, followedUsername string) error
}
