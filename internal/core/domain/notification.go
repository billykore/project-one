package domain

import (
	"errors"
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
	ID        int              `json:"id"`
	UserID    int              `json:"user_id"`
	ActorID   int              `json:"actor_id"`
	Type      NotificationType `json:"type"`
	PostID    *int             `json:"post_id,omitempty"`
	CommentID *int             `json:"comment_id,omitempty"`
	IsRead    bool             `json:"is_read"`
	CreatedAt time.Time        `json:"created_at"`
}

type NotificationDetail struct {
	Notification
	ActorUsername string `json:"actor_username"`
}

func (n *Notification) Validate() error {
	if n.UserID <= 0 {
		return ErrInvalidNotification
	}
	if n.ActorID <= 0 {
		return ErrInvalidNotification
	}
	if n.Type != NotificationTypeFollow && n.Type != NotificationTypeLike && n.Type != NotificationTypeComment {
		return ErrInvalidNotification
	}
	return nil
}
