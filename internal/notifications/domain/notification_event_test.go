package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationEventMarshalJSON(t *testing.T) {
	createdAt := time.Date(2026, 7, 25, 10, 30, 0, 123000000, time.UTC)
	event := NotificationEvent{
		EventID:       "backend-123",
		SchemaVersion: "1.0",
		Timestamp:     "2026-07-25T10:30:00.123Z",
		Notification: Notification{
			Type:          NotificationTypeComment,
			UserID:        42,
			ActorID:       7,
			ActorUsername: "alice",
			PostID:        108,
			CommentID:     512,
			CreatedAt:     createdAt,
		},
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)
	assert.JSONEq(t, `{"event_id":"backend-123","schema_version":"1.0","type":"comment","user_id":42,"actor_id":7,"actor_username":"alice","post_id":108,"comment_id":512,"created_at":"2026-07-25T10:30:00.123Z","timestamp":"2026-07-25T10:30:00.123Z"}`, string(data))
}

func TestNotificationEventUnmarshalJSON(t *testing.T) {
	input := []byte(`{"event_id":"backend-123","schema_version":"1.0","type":"like","user_id":42,"actor_id":7,"actor_username":"alice","post_id":108,"comment_id":0,"created_at":"2026-07-25T10:30:00.123Z","timestamp":"2026-07-25T10:30:00.123Z"}`)

	var event NotificationEvent
	require.NoError(t, json.Unmarshal(input, &event))
	assert.Equal(t, "backend-123", event.EventID)
	assert.Equal(t, "1.0", event.SchemaVersion)
	assert.Equal(t, NotificationTypeLike, event.Notification.Type)
	assert.Equal(t, 42, event.Notification.UserID)
	assert.Equal(t, 7, event.Notification.ActorID)
	assert.Equal(t, "alice", event.Notification.ActorUsername)
	assert.Equal(t, 108, event.Notification.PostID)
	assert.Equal(t, 0, event.Notification.CommentID)
	assert.Equal(t, "2026-07-25T10:30:00.123Z", event.Timestamp)
}
