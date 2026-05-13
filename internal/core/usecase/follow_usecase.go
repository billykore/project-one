package usecase

import (
	"context"
	"fmt"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type followUseCase struct {
	followRepo ports.FollowRepository
	userRepo   ports.UserRepository
}

// NewFollowUseCase creates a new instance of FollowUseCase.
func NewFollowUseCase(followRepo ports.FollowRepository, userRepo ports.UserRepository) ports.FollowUseCase {
	if followRepo == nil || userRepo == nil {
		panic("NewFollowUseCase: dependencies must not be nil")
	}
	return &followUseCase{
		followRepo: followRepo,
		userRepo:   userRepo,
	}
}

func (u *followUseCase) Follow(ctx context.Context, followerID, followedID int) (*domain.Follow, error) {
	if followerID == followedID {
		return nil, domain.ErrCannotFollowSelf
	}

	// Verify followed user exists
	_, err := u.userRepo.GetUserByID(ctx, followedID)
	if err != nil {
		return nil, fmt.Errorf("verify followed user: %w", err)
	}

	// Check if already following
	isFollowing, err := u.followRepo.IsFollowing(ctx, followerID, followedID)
	if err != nil {
		return nil, fmt.Errorf("check following status: %w", err)
	}
	if isFollowing {
		return nil, domain.ErrAlreadyFollowing
	}

	follow := &domain.Follow{
		FollowerID: followerID,
		FollowedID: followedID,
	}

	if err := u.followRepo.Create(ctx, follow); err != nil {
		return nil, fmt.Errorf("create follow: %w", err)
	}

	return follow, nil
}
