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
)
