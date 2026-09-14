package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	vo "github.com/billykore/project-one/internal/core/valueobject"
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

	notificationEvent := domain.NotificationEvent{
		EventID:       fmt.Sprintf("backend-%d", time.Now().UnixNano()),
		SchemaVersion: "1.0",
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Notification:  *notification,
	}

	payload, err := json.Marshal(notificationEvent)
	if err != nil {
		u.log.Error(ctx, "failed to marshal follow notification", "error", err)
		return follow, nil
	}

	event := ports.Event{
		Topic:   followNotificationTopic,
		Key:     fmt.Sprintf("user:%d", followed.ID),
		Payload: payload,
		Metadata: map[string]string{
			"event_id":       notificationEvent.EventID,
			"schema_version": notificationEvent.SchemaVersion,
			"timestamp":      notificationEvent.Timestamp,
		},
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

func (u *followUseCase) GetFollowing(ctx context.Context, followerUsername string, cursor *vo.Cursor, limit int) (*ports.FollowingPage, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	follower, err := u.userRepo.GetUserByUsername(ctx, followerUsername)
	if err != nil {
		return nil, fmt.Errorf("get follower by username: %w", err)
	}
	if follower == nil {
		return nil, domain.ErrUserNotFound
	}

	following, err := u.followRepo.GetFollowing(ctx, follower.Username, cursor, limit+1)
	if err != nil {
		return nil, fmt.Errorf("get following: %w", err)
	}

	page := &ports.FollowingPage{Data: following}
	if len(following) > limit {
		page.HasMore = true
		page.Data = following[:limit]
		last := page.Data[len(page.Data)-1]
		page.NextCursor = &vo.Cursor{CreatedAt: last.FollowedAt, Key: last.Username}
	}
	return page, nil
}

func (u *followUseCase) GetFollowers(ctx context.Context, followedUsername string, cursor *vo.Cursor, limit int) (*ports.FollowersPage, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	followed, err := u.userRepo.GetUserByUsername(ctx, followedUsername)
	if err != nil {
		return nil, fmt.Errorf("get followed by username: %w", err)
	}
	if followed == nil {
		return nil, domain.ErrUserNotFound
	}

	followers, err := u.followRepo.GetFollowers(ctx, followed.Username, cursor, limit+1)
	if err != nil {
		return nil, fmt.Errorf("get followers: %w", err)
	}

	page := &ports.FollowersPage{Data: followers}
	if len(followers) > limit {
		page.HasMore = true
		page.Data = followers[:limit]
		last := page.Data[len(page.Data)-1]
		page.NextCursor = &vo.Cursor{CreatedAt: last.FollowedAt, Key: last.Username}
	}
	return page, nil
}
