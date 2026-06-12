package usecase

import (
	"context"

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
		return nil, err
	}
	notifications, err := uc.repo.GetByUserID(ctx, user.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	actorMap := make(map[int]string)
	details := make([]*domain.NotificationDetail, len(notifications))
	for i, n := range notifications {
		actorUsername, exists := actorMap[n.ActorID]
		if !exists {
			actor, err := uc.userRepo.GetUserByID(ctx, n.ActorID)
			if err == nil {
				actorUsername = actor.Username
				actorMap[n.ActorID] = actorUsername
			}
		}
		details[i] = &domain.NotificationDetail{
			Notification:  *n,
			ActorUsername: actorUsername,
		}
	}
	return details, nil
}

func (uc *notificationUseCase) MarkAsRead(ctx context.Context, id int, username string) error {
	user, err := uc.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	notification, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if notification.UserID != user.ID {
		return domain.ErrUnauthorized
	}
	return uc.repo.MarkAsRead(ctx, id)
}

func (uc *notificationUseCase) MarkAllAsRead(ctx context.Context, username string) error {
	user, err := uc.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	return uc.repo.MarkAllAsRead(ctx, user.ID)
}
