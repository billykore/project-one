package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type notificationUseCase struct {
	repo     ports.NotificationRepository
	userRepo ports.UserRepository
}

func NewNotificationUseCase(repo ports.NotificationRepository, userRepo ports.UserRepository) ports.NotificationUseCase {
	if repo == nil {
		panic("NewNotificationUseCase: repo is required")
	}
	if userRepo == nil {
		panic("NewNotificationUseCase: userRepo is required")
	}
	return &notificationUseCase{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (uc *notificationUseCase) GetNotifications(ctx context.Context, username string, limit, offset int) ([]*domain.NotificationDetail, error) {
	user, err := uc.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("get user by username: %w", domain.ErrUserNotFound)
	}
	notifications, err := uc.repo.GetByUserID(ctx, user.ID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get notifications by user id: %w", err)
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
				if errors.Is(err, domain.ErrUserNotFound) {
					actorUsername = ""
					actorMap[n.ActorID] = ""
				} else {
					return nil, fmt.Errorf("get actor by id: %w", err)
				}
			} else {
				if actor != nil {
					actorUsername = actor.Username
				} else {
					actorUsername = ""
				}
				actorMap[n.ActorID] = actorUsername
			}
		}
		details = append(details, &domain.NotificationDetail{
			Notification:  *n,
			ActorUsername: actorUsername,
		})
	}
	return details, nil
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
		return domain.ErrUnauthorized
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
