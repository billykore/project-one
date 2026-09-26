package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// NotificationEvent represents a notification event published to the message broker.
// It wraps the core Notification with broker-specific metadata for idempotency,
// schema versioning, and event timing.
type NotificationEvent struct {
	// EventID is a unique identifier (format: "backend-<uuid>") used for consumer idempotency.
	EventID string `json:"event_id"`
	// SchemaVersion indicates the event schema version for consumer compatibility.
	SchemaVersion string `json:"schema_version"`
	// Timestamp is the ISO 8601 publication timestamp.
	Timestamp string `json:"timestamp"`
	// Notification contains the notification payload.
	Notification Notification `json:"notification"`
}

type notificationEventJSON struct {
	EventID       string           `json:"event_id"`
	SchemaVersion string           `json:"schema_version"`
	Type          NotificationType `json:"type"`
	UserID        int              `json:"user_id"`
	ActorID       int              `json:"actor_id"`
	ActorUsername string           `json:"actor_username"`
	PostID        int              `json:"post_id"`
	CommentID     int              `json:"comment_id"`
	CreatedAt     string           `json:"created_at"`
	Timestamp     string           `json:"timestamp"`
}

// MarshalJSON flattens the notification event to match the broker contract.
func (e NotificationEvent) MarshalJSON() ([]byte, error) {
	payload := notificationEventJSON{
		EventID:       e.EventID,
		SchemaVersion: e.SchemaVersion,
		Type:          e.Notification.Type,
		UserID:        e.Notification.UserID,
		ActorID:       e.Notification.ActorID,
		ActorUsername: e.Notification.ActorUsername,
		PostID:        e.Notification.PostID,
		CommentID:     e.Notification.CommentID,
		CreatedAt:     e.Notification.CreatedAt.UTC().Format(time.RFC3339Nano),
		Timestamp:     e.Timestamp,
	}
	return json.Marshal(payload)
}

// UnmarshalJSON restores the flattened broker payload into the nested domain type.
func (e *NotificationEvent) UnmarshalJSON(data []byte) error {
	var payload notificationEventJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	if payload.EventID == "" {
		return fmt.Errorf("notification event_id cannot be empty")
	}

	createdAt, err := time.Parse(time.RFC3339Nano, payload.CreatedAt)
	if err != nil {
		return fmt.Errorf("notification created_at: %w", err)
	}

	e.EventID = payload.EventID
	e.SchemaVersion = payload.SchemaVersion
	e.Timestamp = payload.Timestamp
	e.Notification = Notification{
		Type:          payload.Type,
		UserID:        payload.UserID,
		ActorID:       payload.ActorID,
		ActorUsername: payload.ActorUsername,
		PostID:        payload.PostID,
		CommentID:     payload.CommentID,
		CreatedAt:     createdAt,
	}
	return nil
}
