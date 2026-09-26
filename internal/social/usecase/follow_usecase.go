package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	identitydomain "github.com/billykore/project-one/internal/identity/domain"
	notificationdomain "github.com/billykore/project-one/internal/notifications/domain"
	vo "github.com/billykore/project-one/internal/platform/pagination"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	"github.com/billykore/project-one/internal/platform/problem"
	"github.com/billykore/project-one/internal/social/domain"
	"github.com/billykore/project-one/internal/social/ports"
)

const followNotificationTopic = "notifications"

type followUseCase struct {
	followRepo ports.FollowRepository
	userRepo   ports.AccountLookup
	publisher  platformports.Publisher
	log        platformports.Logger
}

// NewFollowUseCase creates a new instance of FollowUseCase.
func NewFollowUseCase(
	followRepo ports.FollowRepository,
	userRepo ports.AccountLookup,
	publisher platformports.Publisher,
	log platformports.Logger,
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

func (u *followUseCase) Follow(ctx context.Context, actor *identitydomain.User, followedUsername string) (*domain.Follow, error) {
	if actor == nil || actor.ID <= 0 {
		return nil, problem.ErrInvalidUser
	}

	followed, err := u.userRepo.GetUserByUsername(ctx, followedUsername)
	if err != nil {
		return nil, fmt.Errorf("get followed by username: %w", err)
	}

	if followed == nil {
		return nil, fmt.Errorf("get user: %w", problem.ErrUserNotFound)
	}
	if actor.ID == followed.ID {
		return nil, problem.ErrCannotFollowSelf
	}

	follow := &domain.Follow{
		FollowerID:       actor.ID,
		FollowerUsername: actor.Username,
		FollowedID:       followed.ID,
		FollowedUsername: followed.Username,
	}

	if err := u.followRepo.Create(ctx, follow); err != nil {
		return nil, fmt.Errorf("create follow: %w", err)
	}

	notification := &notificationdomain.Notification{
		UserID:        followed.ID,
		ActorID:       actor.ID,
		ActorUsername: actor.Username,
		Type:          notificationdomain.NotificationTypeFollow,
		CreatedAt:     follow.CreatedAt,
	}

	notificationEvent := notificationdomain.NotificationEvent{
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

	event := platformports.Event{
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

func (u *followUseCase) Unfollow(ctx context.Context, actor *identitydomain.User, followedUsername string) error {
	if actor == nil || actor.ID <= 0 {
		return problem.ErrInvalidUser
	}
	followed, err := u.userRepo.GetUserByUsername(ctx, followedUsername)
	if err != nil {
		return fmt.Errorf("get followed by username: %w", err)
	}
	if followed == nil {
		return problem.ErrUserNotFound
	}
	if actor.ID == followed.ID {
		return problem.ErrCannotUnfollowSelf
	}
	if err := u.followRepo.Delete(ctx, actor.ID, followed.ID); err != nil {
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
		return nil, problem.ErrUserNotFound
	}

	following, err := u.followRepo.GetFollowing(ctx, follower.ID, cursor, limit+1)
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
		return nil, problem.ErrUserNotFound
	}

	followers, err := u.followRepo.GetFollowers(ctx, followed.ID, cursor, limit+1)
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
