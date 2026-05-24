package domain

import (
	"errors"
	"time"
)

var (
	// ErrAlreadyLiked is returned when a user tries to like a post they already liked.
	ErrAlreadyLiked = errors.New("post already liked")
	// ErrNotLiked is returned when a user tries to unlike a post they haven't liked.
	ErrNotLiked = errors.New("post not liked")
)

// Like is the core domain entity representing a user's like on a post.
type Like struct {
	PostID    int
	Username  string
	CreatedAt time.Time
}
