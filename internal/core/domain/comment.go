package domain

import (
	"fmt"
	"strings"
	"time"
)

// Comment is the core domain entity representing a comment on a post.
type Comment struct {
	ID        int64
	PostID    int64
	Username  string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

// Validate performs domain-level validation on the Comment entity.
func (c *Comment) Validate() error {
	if len(strings.TrimSpace(c.Content)) < 1 {
		return fmt.Errorf("%w: comment must be at least 1 character", ErrValidationFailed)
	}
	return nil
}
