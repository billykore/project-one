package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
)

// NotificationRepository is a driven port for notification database persistence.
type NotificationRepository interface {
	// Create saves a new notification in the database.
	Create(ctx context.Context, notification *domain.Notification) error
	// GetByID retrieves a single notification by its unique identifier.
	GetByID(ctx context.Context, id int) (*domain.Notification, error)
	// GetByUserID retrieves notifications for a specific user with pagination support, sorted descending by creation time.
	GetByUserID(ctx context.Context, userID int, limit, offset int) ([]*domain.Notification, error)
	// MarkAsRead marks a specific notification as read.
	MarkAsRead(ctx context.Context, id int) error
	// MarkAllAsRead marks all notifications as read for a given user.
	MarkAllAsRead(ctx context.Context, userID int) error
}

// NotificationUseCase is a driving port for notification business logic.
type NotificationUseCase interface {
	// GetNotifications retrieves notifications for the given user, resolved with actor details.
	GetNotifications(ctx context.Context, username string, limit, offset int) ([]*domain.NotificationDetail, error)
	// MarkAsRead verifies ownership and marks a specific notification as read.
	MarkAsRead(ctx context.Context, id int, username string) error
	// MarkAllAsRead marks all notifications as read for the authenticated user.
	MarkAllAsRead(ctx context.Context, username string) error
	// SaveNotification saves a notification to the database.
	SaveNotification(ctx context.Context, notification *domain.Notification) error
}

// NotificationPublisher is a driven port for publishing notifications to an event broker asynchronously.
type NotificationPublisher interface {
	// Publish sends a notification event to the message broker.
	Publish(ctx context.Context, notification *domain.Notification) error
}

// NotificationConsumer is a driver port representing a background worker that consumes notification events.
type NotificationConsumer interface {
	// Start begins the asynchronous consumption of events from the broker.
	Start(ctx context.Context) (<-chan *domain.Notification, error)
	// Stop gracefully terminates the background worker process.
	Stop(ctx context.Context) error
}
