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
	GetFollowing(ctx context.Context, followerID int, cursor *vo.Cursor, limit int) ([]domain.Following, error)
	// GetFollowers fetches the cursor-paginated list of users following a specific user.
	GetFollowers(ctx context.Context, followedID int, cursor *vo.Cursor, limit int) ([]domain.Follower, error)
	// Delete removes an existing follow relationship.
	Delete(ctx context.Context, followerID, followedID int) error
	// GetFollowedUserIDs returns stable IDs for every user the actor follows.
	GetFollowedUserIDs(ctx context.Context, followerID int) ([]int, error)
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
	// Follow creates a relationship from the authenticated actor to a user selected by username.
	Follow(ctx context.Context, actor *domain.User, followedUsername string) (*domain.Follow, error)
	// GetFollowing handles the logic for getting a cursor-paginated following list.
	GetFollowing(ctx context.Context, followerUsername string, cursor *vo.Cursor, limit int) (*FollowingPage, error)
	// GetFollowers handles the logic for getting a cursor-paginated followers list.
	GetFollowers(ctx context.Context, followedUsername string, cursor *vo.Cursor, limit int) (*FollowersPage, error)
	// Unfollow removes a relationship from the authenticated actor to a user selected by username.
	Unfollow(ctx context.Context, actor *domain.User, followedUsername string) error
}
