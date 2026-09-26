package dto

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
	ID        int                `json:"id"`
	Message   string             `json:"message,omitempty"`
	Title     string             `json:"title,omitempty"`
	Content   string             `json:"content,omitempty"`
	Tags      []string           `json:"tags,omitempty"`
	Author    string             `json:"author,omitempty"`
	LikeCount int                `json:"like_count"`
	Comments  []*CommentResponse `json:"comments,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type UpdatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
