package sse

import (
	"context"
	"testing"

	"github.com/billykore/project-one/internal/notifications/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifier_ReportsOfflineRecipientWithoutError(t *testing.T) {
	delivered, err := NewNotifier(NewManager()).Notify(context.Background(), domain.Notification{UserID: 42})
	assert.NoError(t, err)
	assert.False(t, delivered)
}

func TestNotifier_DeliversToConnectedRecipient(t *testing.T) {
	manager := NewManager()
	ch := manager.Register(42)
	defer manager.Unregister(42, ch)

	delivered, err := NewNotifier(manager).Notify(context.Background(), domain.Notification{UserID: 42, ActorUsername: "bob", Type: domain.NotificationTypeFollow})
	require.NoError(t, err)
	assert.True(t, delivered)

	got := <-ch
	assert.Equal(t, 42, got.UserID)
	assert.Equal(t, "bob", got.ActorUsername)
}
