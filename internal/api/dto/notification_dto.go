package dto

import (
	"time"

	"github.com/billykore/project-one/internal/core/domain"
)

type NotificationResponse struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	ActorID       int       `json:"actor_id"`
	ActorUsername string    `json:"actor_username"`
	Type          string    `json:"type"`
	PostID        int       `json:"post_id,omitempty"`
	CommentID     int       `json:"comment_id,omitempty"`
	IsRead        bool      `json:"is_read"`
	CreatedAt     time.Time `json:"created_at"`
	Title         string    `json:"title,omitempty"`
	Body          string    `json:"body,omitempty"`
}

func NotificationTitle(notificationType domain.NotificationType) string {
	switch notificationType {
	case domain.NotificationTypeFollow:
		return "New Follower"
	case domain.NotificationTypeLike:
		return "New Like"
	case domain.NotificationTypeComment:
		return "New Comment"
	default:
		return "Notification"
	}
}

func NotificationBody(notificationType domain.NotificationType, actorUsername string) string {
	switch notificationType {
	case domain.NotificationTypeFollow:
		return actorUsername + " started following you."
	case domain.NotificationTypeLike:
		return actorUsername + " liked your post."
	case domain.NotificationTypeComment:
		return actorUsername + " commented on your post."
	default:
		return ""
	}
}
