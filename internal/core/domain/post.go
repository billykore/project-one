package domain

import (
	"time"
)

// Post is the core domain entity representing a user's post.
type Post struct {
	ID        int
	Username  string
	Title     string
	Content   string
	Tags      []string
	LikeCount int
	CreatedAt time.Time
	UpdatedAt time.Time
}
