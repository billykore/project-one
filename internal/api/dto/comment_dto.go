package dto

import "time"

type CreateCommentRequest struct {
	ID      int    `param:"id" validate:"required,min=1"`
	Content string `json:"content" validate:"required,min=1"`
}

type CommentResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
