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
	Type      NotificationType
	ID        int
	UserID    int
	ActorID   int
	PostID    int
	CommentID int
	IsRead    bool
	CreatedAt time.Time
}

type NotificationDetail struct {
	Notification
	ActorUsername string
}

func (n *Notification) Validate() error {
	return nil
}
