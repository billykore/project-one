package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNotification_Validate(t *testing.T) {
	t.Run("valid notification", func(t *testing.T) {
		n := &Notification{
			UserID:  1,
			ActorID: 2,
			Type:    NotificationTypeFollow,
		}
		assert.NoError(t, n.Validate())
	})

	t.Run("missing user id", func(t *testing.T) {
		n := &Notification{
			ActorID: 2,
			Type:    NotificationTypeFollow,
		}
		assert.ErrorIs(t, n.Validate(), ErrInvalidNotification)
	})

	t.Run("invalid type", func(t *testing.T) {
		n := &Notification{
			UserID:  1,
			ActorID: 2,
			Type:    NotificationType("invalid"),
		}
		assert.ErrorIs(t, n.Validate(), ErrInvalidNotification)
	})
}
