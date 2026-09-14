package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
	vo "github.com/billykore/project-one/internal/core/valueobject"
)

// NotificationRepository is a driven port for notification database persistence.
type NotificationRepository interface {
	// Create saves a new notification in the database.
	Create(ctx context.Context, notification *domain.Notification) error
	// GetByID retrieves a single notification by its unique identifier.
	GetByID(ctx context.Context, id int) (*domain.Notification, error)
	// GetByUserID retrieves notifications for a specific user after cursor, newest first.
	GetByUserID(ctx context.Context, userID int, cursor *vo.Cursor, limit int) ([]*domain.Notification, error)
	// MarkAsRead marks a specific notification as read.
	MarkAsRead(ctx context.Context, id int) error
	// MarkAllAsRead marks all notifications as read for a given user.
	MarkAllAsRead(ctx context.Context, userID int) error
}

// NotificationsPage is a cursor-paginated notification result.
type NotificationsPage struct {
	Notifications []*domain.NotificationDetail
	NextCursor    *vo.Cursor
	HasMore       bool
}

// NotificationUseCase is a driving port for notification business logic.
type NotificationUseCase interface {
	// GetNotifications retrieves a cursor-paginated notification page with actor details.
	GetNotifications(ctx context.Context, username string, cursor *vo.Cursor, limit int) (*NotificationsPage, error)
	// MarkAsRead verifies ownership and marks a specific notification as read.
	MarkAsRead(ctx context.Context, id int, username string) error
	// MarkAllAsRead marks all notifications as read for the authenticated user.
	MarkAllAsRead(ctx context.Context, username string) error
	// SaveNotification saves a notification to the database.
	SaveNotification(ctx context.Context, notification *domain.Notification) error
}
