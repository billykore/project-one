package domain

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrInvalidNotification  = errors.New("invalid notification")
)

type NotificationType string

const (
	NotificationTypeFollow  NotificationType = "follow"
	NotificationTypeLike    NotificationType = "like"
	NotificationTypeComment NotificationType = "comment"
)

type Notification struct {
	ID        int
	UserID    int
	ActorID   int
	Type      NotificationType
	PostID    *int
	CommentID *int
	IsRead    bool
	CreatedAt time.Time
}

type NotificationDetail struct {
	Notification
	ActorUsername string
}

func (n *Notification) Validate() error {
	if n.UserID <= 0 {
		return fmt.Errorf("%w: user id is required", ErrValidationFailed)
	}
	if n.ActorID <= 0 {
		return fmt.Errorf("%w: actor id is required", ErrValidationFailed)
	}
	if n.Type != NotificationTypeFollow && n.Type != NotificationTypeLike && n.Type != NotificationTypeComment {
		return fmt.Errorf("%w: invalid notification type", ErrValidationFailed)
	}

	switch n.Type {
	case NotificationTypeFollow:
		if n.PostID != nil {
			return fmt.Errorf("%w: post id must be nil for follow type", ErrValidationFailed)
		}
		if n.CommentID != nil {
			return fmt.Errorf("%w: comment id must be nil for follow type", ErrValidationFailed)
		}
	case NotificationTypeLike:
		if n.PostID == nil {
			return fmt.Errorf("%w: post id is required for like type", ErrValidationFailed)
		}
		if *n.PostID <= 0 {
			return fmt.Errorf("%w: invalid post id for like type", ErrValidationFailed)
		}
		if n.CommentID != nil {
			return fmt.Errorf("%w: comment id must be nil for like type", ErrValidationFailed)
		}
	case NotificationTypeComment:
		if n.PostID == nil {
			return fmt.Errorf("%w: post id is required for comment type", ErrValidationFailed)
		}
		if *n.PostID <= 0 {
			return fmt.Errorf("%w: invalid post id for comment type", ErrValidationFailed)
		}
		if n.CommentID == nil {
			return fmt.Errorf("%w: comment id is required for comment type", ErrValidationFailed)
		}
		if *n.CommentID <= 0 {
			return fmt.Errorf("%w: invalid comment id for comment type", ErrValidationFailed)
		}
	}

	return nil
}
