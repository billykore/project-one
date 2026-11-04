package domain

import "errors"

// Sentinel errors owned by the notifications context.
var (
	// ErrNotificationNotFound is returned when a notification cannot be found in the system.
	ErrNotificationNotFound = errors.New("notification not found")
	// ErrNotificationNotOwned is returned when a user tries to access or modify a notification they do not own.
	ErrNotificationNotOwned = errors.New("notification not owned by user")
	// ErrInvalidNotification is returned when a notification is invalid.
	ErrInvalidNotification = errors.New("invalid notification")
	// ErrDuplicateNotification is returned when a notification for the same event was already stored.
	ErrDuplicateNotification = errors.New("duplicate notification event")
	// ErrUnsupportedEventSchema is returned when a notification event uses an unknown schema version.
	ErrUnsupportedEventSchema = errors.New("unsupported notification event schema version")
)
