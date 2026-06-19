package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

const followNotificationTopic = "notifications"

type followUseCase struct {
	followRepo ports.FollowRepository
	userRepo   ports.UserRepository
	publisher  ports.Publisher
	log        ports.Logger
}

// NewFollowUseCase creates a new instance of FollowUseCase.
func NewFollowUseCase(
	followRepo ports.FollowRepository,
	userRepo ports.UserRepository,
	publisher ports.Publisher,
	log ports.Logger,
) ports.FollowUseCase {
	if followRepo == nil || userRepo == nil || publisher == nil || log == nil {
		panic("NewFollowUseCase: dependencies must not be nil")
	}
	return &followUseCase{
		followRepo: followRepo,
		userRepo:   userRepo,
		publisher:  publisher,
		log:        log,
	}
}

func (u *followUseCase) Follow(ctx context.Context, followerUsername, followedUsername string) (*domain.Follow, error) {
	if followerUsername == followedUsername {
		return nil, domain.ErrCannotFollowSelf
	}

	follower, err := u.userRepo.GetUserByUsername(ctx, followerUsername)
	if err != nil {
		return nil, fmt.Errorf("get follower by username: %w", err)
	}

	followed, err := u.userRepo.GetUserByUsername(ctx, followedUsername)
	if err != nil {
		return nil, fmt.Errorf("get followed by username: %w", err)
	}

	if follower == nil || followed == nil {
		return nil, fmt.Errorf("get user: %w", domain.ErrUserNotFound)
	}

	follow := &domain.Follow{
		FollowerID:       follower.ID,
		FollowerUsername: follower.Username,
		FollowedID:       followed.ID,
		FollowedUsername: followed.Username,
	}

	if err := u.followRepo.Create(ctx, follow); err != nil {
		return nil, fmt.Errorf("create follow: %w", err)
	}

	notification := &domain.Notification{
		UserID:        followed.ID,
		ActorID:       follower.ID,
		ActorUsername: follower.Username,
		Type:          domain.NotificationTypeFollow,
		CreatedAt:     follow.CreatedAt,
	}
	if err := notification.Validate(); err != nil {
		u.log.Error(ctx, "invalid follow notification", "error", err)
		return follow, nil
	}

	payload, err := json.Marshal(notification)
	if err != nil {
		u.log.Error(ctx, "failed to marshal follow notification", "error", err)
		return follow, nil
	}

	event := ports.Event{
		Topic:   followNotificationTopic,
		Key:     fmt.Sprintf("user:%d", followed.ID),
		Payload: payload,
	}
	if err := u.publisher.Publish(ctx, event); err != nil {
		u.log.Error(ctx, "failed to publish follow notification", "error", err)
	}

	return follow, nil
}

func (u *followUseCase) Unfollow(ctx context.Context, followerUsername, followedUsername string) error {
	if followerUsername == followedUsername {
		return domain.ErrCannotUnfollowSelf
	}

	if err := u.followRepo.Delete(ctx, followerUsername, followedUsername); err != nil {
		return fmt.Errorf("delete follow: %w", err)
	}

	return nil
}

func (u *followUseCase) GetFollowing(ctx context.Context, followerUsername string, limit, offset int) ([]domain.Following, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	follower, err := u.userRepo.GetUserByUsername(ctx, followerUsername)
	if err != nil {
		return nil, fmt.Errorf("get follower by username: %w", err)
	}
	if follower == nil {
		return nil, domain.ErrUserNotFound
	}

	following, err := u.followRepo.GetFollowing(ctx, follower.Username, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get following: %w", err)
	}

	return following, nil
}

func (u *followUseCase) GetFollowers(ctx context.Context, followedUsername string, limit, offset int) ([]domain.Follower, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	followed, err := u.userRepo.GetUserByUsername(ctx, followedUsername)
	if err != nil {
		return nil, fmt.Errorf("get followed by username: %w", err)
	}
	if followed == nil {
		return nil, domain.ErrUserNotFound
	}

	followers, err := u.followRepo.GetFollowers(ctx, followed.Username, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get followers: %w", err)
	}

	return followers, nil
}
