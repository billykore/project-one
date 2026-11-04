package ports

import (
	"context"

	"github.com/billykore/project-one/internal/notifications/domain"
	"github.com/billykore/project-one/internal/pkg/auth"
	vo "github.com/billykore/project-one/internal/pkg/pagination"
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
	// GetNotifications retrieves a cursor-paginated notification page for an authenticated recipient.
	GetNotifications(ctx context.Context, recipient *auth.Principal, cursor *vo.Cursor, limit int) (*NotificationsPage, error)
	// MarkAsRead verifies ownership and marks a specific notification as read.
	MarkAsRead(ctx context.Context, id int, userID int) error
	// MarkAllAsRead marks all notifications as read for the authenticated user.
	MarkAllAsRead(ctx context.Context, userID int) error
	// RecordEvent stores a notification event exactly once and pushes it to the
	// recipient's realtime connections. A redelivered event is a no-op.
	RecordEvent(ctx context.Context, event domain.NotificationEvent) error
}

// RealtimeNotifier is a driven port that pushes a stored notification to the
// recipient's live connections.
type RealtimeNotifier interface {
	// Notify reports whether any live connection received the notification.
	// A recipient without a live connection is not an error.
	Notify(ctx context.Context, notification domain.Notification) (delivered bool, err error)
}
