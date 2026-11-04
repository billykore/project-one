package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/billykore/project-one/internal/notifications/domain"
	"github.com/billykore/project-one/internal/notifications/ports"
	"github.com/billykore/project-one/internal/pkg/auth"
	vo "github.com/billykore/project-one/internal/pkg/pagination"
	pkgports "github.com/billykore/project-one/internal/pkg/ports"
	"github.com/billykore/project-one/internal/pkg/problem"
)

type notificationUseCase struct {
	repo     ports.NotificationRepository
	userRepo ports.ActorLookup
	notifier ports.RealtimeNotifier
	log      pkgports.Logger
}

func NewNotificationUseCase(
	repo ports.NotificationRepository,
	userRepo ports.ActorLookup,
	notifier ports.RealtimeNotifier,
	log pkgports.Logger,
) ports.NotificationUseCase {
	if repo == nil || userRepo == nil || notifier == nil || log == nil {
		panic("NewNotificationUseCase: dependencies must not be nil")
	}
	return &notificationUseCase{
		repo:     repo,
		userRepo: userRepo,
		notifier: notifier,
		log:      log,
	}
}

func (uc *notificationUseCase) GetNotifications(ctx context.Context, recipient *auth.Principal, cursor *vo.Cursor, limit int) (*ports.NotificationsPage, error) {
	if recipient == nil || recipient.ID <= 0 {
		return nil, problem.ErrInvalidUser
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	notifications, err := uc.repo.GetByUserID(ctx, recipient.ID, cursor, limit+1)
	if err != nil {
		return nil, fmt.Errorf("get notifications by user id: %w", err)
	}
	page := &ports.NotificationsPage{}
	if len(notifications) > limit {
		page.HasMore = true
		notifications = notifications[:limit]
		last := notifications[len(notifications)-1]
		if last != nil {
			page.NextCursor = &vo.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
		}
	}

	actorMap := map[int]string{
		recipient.ID: recipient.Username,
	}
	details := make([]*domain.NotificationDetail, 0, len(notifications))
	for _, n := range notifications {
		if n == nil {
			continue
		}
		actorUsername, exists := actorMap[n.ActorID]
		if !exists {
			actor, err := uc.userRepo.GetUserByID(ctx, n.ActorID)
			if err != nil {
				if !errors.Is(err, problem.ErrUserNotFound) {
					return nil, fmt.Errorf("get actor by id: %w", err)
				}
			} else if actor != nil {
				actorUsername = actor.Username
			}
			actorMap[n.ActorID] = actorUsername
		}
		details = append(details, &domain.NotificationDetail{
			Notification:  *n,
			ActorUsername: actorUsername,
		})
	}
	page.Notifications = details
	return page, nil
}

func (uc *notificationUseCase) MarkAsRead(ctx context.Context, id int, userID int) error {
	if userID <= 0 {
		return problem.ErrInvalidUser
	}

	notification, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get notification by id: %w", err)
	}
	if notification == nil {
		return domain.ErrNotificationNotFound
	}
	if notification.UserID != userID {
		return domain.ErrNotificationNotOwned
	}

	err = uc.repo.MarkAsRead(ctx, id)
	if err != nil {
		return fmt.Errorf("mark notification as read: %w", err)
	}

	return nil
}

func (uc *notificationUseCase) MarkAllAsRead(ctx context.Context, userID int) error {
	if userID <= 0 {
		return problem.ErrInvalidUser
	}
	err := uc.repo.MarkAllAsRead(ctx, userID)
	if err != nil {
		return fmt.Errorf("mark all notifications as read: %w", err)
	}
	return nil
}

func (uc *notificationUseCase) RecordEvent(ctx context.Context, event domain.NotificationEvent) error {
	if event.SchemaVersion != "" && event.SchemaVersion != domain.NotificationEventSchemaVersion {
		return fmt.Errorf("%w: %s", domain.ErrUnsupportedEventSchema, event.SchemaVersion)
	}

	notification := event.Notification
	notification.EventID = event.EventID
	if err := uc.repo.Create(ctx, &notification); err != nil {
		if errors.Is(err, domain.ErrDuplicateNotification) {
			uc.log.Debug(ctx, "duplicate notification event skipped", "event_id", notification.EventID)
			return nil
		}
		return fmt.Errorf("create notification: %w", err)
	}
	uc.log.Info(ctx, "notification saved", "userID", notification.UserID, "actorID", notification.ActorID, "type", notification.Type)

	// The notification is already stored, so a failed push must not cause a redelivery.
	delivered, err := uc.notifier.Notify(ctx, notification)
	if err != nil {
		uc.log.Warn(ctx, "failed to push notification to realtime clients", "userID", notification.UserID, "error", err)
	} else if delivered {
		uc.log.Info(ctx, "notification streamed to realtime clients", "userID", notification.UserID, "type", notification.Type)
	}
	return nil
}
