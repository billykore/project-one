package domain

import (
	"errors"
	"time"
)

var (
	// ErrPostNotFound is returned when a post cannot be found in the system.
	ErrPostNotFound = errors.New("post not found")
	// ErrInvalidPost is returned when post data is invalid.
	ErrInvalidPost = errors.New("invalid post data")
)

// Post is the core domain entity representing a user's post.
type Post struct {
	ID        int
	Username  string
	Title     string
	Content   string
	Tags      []string
	Comments  []*Comment
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
