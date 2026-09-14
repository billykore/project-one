package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	vo "github.com/billykore/project-one/internal/core/valueobject"
)

type notificationUseCase struct {
	repo     ports.NotificationRepository
	userRepo ports.UserRepository
	log      ports.Logger
}

func NewNotificationUseCase(
	repo ports.NotificationRepository,
	userRepo ports.UserRepository,
	log ports.Logger,
) ports.NotificationUseCase {
	if repo == nil || userRepo == nil || log == nil {
		panic("NewNotificationUseCase: dependencies must not be nil")
	}
	return &notificationUseCase{
		repo:     repo,
		userRepo: userRepo,
		log:      log,
	}
}

func (uc *notificationUseCase) GetNotifications(ctx context.Context, username string, cursor *vo.Cursor, limit int) (*ports.NotificationsPage, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	user, err := uc.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("get user by username: %w", domain.ErrUserNotFound)
	}

	notifications, err := uc.repo.GetByUserID(ctx, user.ID, cursor, limit+1)
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
		user.ID: user.Username,
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
				if !errors.Is(err, domain.ErrUserNotFound) {
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

func (uc *notificationUseCase) MarkAsRead(ctx context.Context, id int, username string) error {
	user, err := uc.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("get user by username: %w", err)
	}
	if user == nil {
		return fmt.Errorf("get user by username: %w", domain.ErrUserNotFound)
	}

	notification, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get notification by id: %w", err)
	}
	if notification == nil {
		return domain.ErrNotificationNotFound
	}
	if notification.UserID != user.ID {
		return domain.ErrNotificationNotOwned
	}

	err = uc.repo.MarkAsRead(ctx, id)
	if err != nil {
		return fmt.Errorf("mark notification as read: %w", err)
	}

	return nil
}

func (uc *notificationUseCase) MarkAllAsRead(ctx context.Context, username string) error {
	user, err := uc.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("get user by username: %w", err)
	}
	if user == nil {
		return fmt.Errorf("get user by username: %w", domain.ErrUserNotFound)
	}
	err = uc.repo.MarkAllAsRead(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("mark all notifications as read: %w", err)
	}
	return nil
}

func (uc *notificationUseCase) SaveNotification(ctx context.Context, notification *domain.Notification) error {
	if err := uc.repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}
