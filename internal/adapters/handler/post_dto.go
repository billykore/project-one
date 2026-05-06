package handler

import "time"

type CreatePostRequest struct {
	Title   string   `json:"title" validate:"required"`
	Content string   `json:"content" validate:"required,min=10"`
	Tags    []string `json:"tags"`
}

type CreatePostResponse struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

type PostResponse struct {
	ID        int       `json:"id"`
	Message   string    `json:"message"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
