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

	follow := &domain.Follow{
		FollowerID: followerID,
		FollowedID: followedID,
	}

	if err := u.followRepo.Create(ctx, follow); err != nil {
		return nil, fmt.Errorf("create follow: %w", err)
	}

	return follow, nil
}

func (u *followUseCase) GetFollowing(ctx context.Context, followerID int, limit, offset int) ([]*domain.Following, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	following, err := u.followRepo.GetFollowing(ctx, followerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get following: %w", err)
	}

	return following, nil
}
