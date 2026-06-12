package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNotification_Validate(t *testing.T) {
	validPostID := 10
	invalidPostID := 0
	validCommentID := 20
	invalidCommentID := 0

	t.Run("valid follow notification", func(t *testing.T) {
		n := &Notification{
			UserID:  1,
			ActorID: 2,
			Type:    NotificationTypeFollow,
		}
		assert.NoError(t, n.Validate())
	})

	t.Run("valid like notification", func(t *testing.T) {
		n := &Notification{
			UserID:  1,
			ActorID: 2,
			Type:    NotificationTypeLike,
			PostID:  &validPostID,
		}
		assert.NoError(t, n.Validate())
	})

	t.Run("valid comment notification", func(t *testing.T) {
		n := &Notification{
			UserID:    1,
			ActorID:   2,
			Type:      NotificationTypeComment,
			PostID:    &validPostID,
			CommentID: &validCommentID,
		}
		assert.NoError(t, n.Validate())
	})

	t.Run("missing user id", func(t *testing.T) {
		n := &Notification{
			ActorID: 2,
			Type:    NotificationTypeFollow,
		}
		assert.ErrorIs(t, n.Validate(), ErrValidationFailed)
	})

	t.Run("missing actor id", func(t *testing.T) {
		n := &Notification{
			UserID: 1,
			Type:   NotificationTypeFollow,
		}
		assert.ErrorIs(t, n.Validate(), ErrValidationFailed)
	})

	t.Run("invalid type", func(t *testing.T) {
		n := &Notification{
			UserID:  1,
			ActorID: 2,
			Type:    NotificationType("invalid"),
		}
		assert.ErrorIs(t, n.Validate(), ErrValidationFailed)
	})

	t.Run("follow with post id", func(t *testing.T) {
		n := &Notification{
			UserID:  1,
			ActorID: 2,
			Type:    NotificationTypeFollow,
			PostID:  &validPostID,
		}
		assert.ErrorIs(t, n.Validate(), ErrValidationFailed)
	})

	t.Run("follow with comment id", func(t *testing.T) {
		n := &Notification{
			UserID:    1,
			ActorID:   2,
			Type:      NotificationTypeFollow,
			CommentID: &validCommentID,
		}
		assert.ErrorIs(t, n.Validate(), ErrValidationFailed)
	})

	t.Run("like with nil post id", func(t *testing.T) {
		n := &Notification{
			UserID:  1,
			ActorID: 2,
			Type:    NotificationTypeLike,
			PostID:  nil,
		}
		assert.ErrorIs(t, n.Validate(), ErrValidationFailed)
	})

	t.Run("like with invalid post id", func(t *testing.T) {
		n := &Notification{
			UserID:  1,
			ActorID: 2,
			Type:    NotificationTypeLike,
			PostID:  &invalidPostID,
		}
		assert.ErrorIs(t, n.Validate(), ErrValidationFailed)
	})

	t.Run("like with comment id", func(t *testing.T) {
		n := &Notification{
			UserID:    1,
			ActorID:   2,
			Type:      NotificationTypeLike,
			PostID:    &validPostID,
			CommentID: &validCommentID,
		}
		assert.ErrorIs(t, n.Validate(), ErrValidationFailed)
	})

	t.Run("comment with nil post id", func(t *testing.T) {
		n := &Notification{
			UserID:    1,
			ActorID:   2,
			Type:      NotificationTypeComment,
			PostID:    nil,
			CommentID: &validCommentID,
		}
		assert.ErrorIs(t, n.Validate(), ErrValidationFailed)
	})

	t.Run("comment with invalid post id", func(t *testing.T) {
		n := &Notification{
			UserID:    1,
			ActorID:   2,
			Type:      NotificationTypeComment,
			PostID:    &invalidPostID,
			CommentID: &validCommentID,
		}
		assert.ErrorIs(t, n.Validate(), ErrValidationFailed)
	})

	t.Run("comment with nil comment id", func(t *testing.T) {
		n := &Notification{
			UserID:    1,
			ActorID:   2,
			Type:      NotificationTypeComment,
			PostID:    &validPostID,
			CommentID: nil,
		}
		assert.ErrorIs(t, n.Validate(), ErrValidationFailed)
	})

	t.Run("comment with invalid comment id", func(t *testing.T) {
		n := &Notification{
			UserID:    1,
			ActorID:   2,
			Type:      NotificationTypeComment,
			PostID:    &validPostID,
			CommentID: &invalidCommentID,
		}
		assert.ErrorIs(t, n.Validate(), ErrValidationFailed)
	})
}
