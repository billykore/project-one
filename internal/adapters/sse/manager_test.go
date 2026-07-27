package sse

import (
	"testing"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/stretchr/testify/assert"
)

func TestManager_RegisterSendAndUnregister(t *testing.T) {
	m := NewManager()

	msgCh := m.Register(10)
	assert.Equal(t, 1, m.ConnectionCount(10))

	err := m.Send(&dto.NotificationResponse{UserID: 10, Type: "follow"})
	assert.NoError(t, err)

	select {
	case msg := <-msgCh:
		assert.Equal(t, 10, msg.UserID)
		assert.Equal(t, "follow", msg.Type)
	default:
		t.Fatal("expected notification")
	}

	m.Unregister(10, msgCh)
	assert.Equal(t, 0, m.ConnectionCount(10))
}

func TestManager_SendToDisconnectedUserReturnsError(t *testing.T) {
	m := NewManager()
	err := m.Send(&dto.NotificationResponse{UserID: 999, Type: "follow"})
	assert.ErrorIs(t, err, ErrUserNotConnected)
}

func TestManager_Close(t *testing.T) {
	m := NewManager()
	msgCh := m.Register(1)

	assert.NoError(t, m.Close())

	_, ok := <-msgCh
	assert.False(t, ok)
	assert.NoError(t, m.Close())
}
