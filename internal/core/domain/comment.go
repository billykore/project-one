package domain

import (
	"strings"
	"time"
)

// Comment is the core domain entity representing a comment on a post.
type Comment struct {
	ID        int
	PostID    int
	Username  string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate performs domain-level validation on the Comment entity.
func (c *Comment) Validate() error {
	if len(strings.TrimSpace(c.Content)) < 1 {
		return ErrCommentTooShort
	}
	return nil
}
