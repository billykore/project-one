package dto

import "time"

type NotificationResponse struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ActorID   int       `json:"actor_id"`
	ActorName string    `json:"actor_name"`
	Type      string    `json:"type"`
	PostID    int       `json:"post_id,omitempty"`
	CommentID int       `json:"comment_id,omitempty"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
