package domain

import (
	"strings"
	"time"

	"github.com/billykore/project-one/internal/platform/problem"
)

// Comment is the core domain entity representing a comment on a post.
type Comment struct {
	ID     int
	PostID int
	UserID int
	// Username is a display projection populated when comments are read.
	Username  string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate performs domain-level validation on the Comment entity.
func (c *Comment) Validate() error {
	if len(strings.TrimSpace(c.Content)) < 1 {
		return problem.ErrCommentTooShort
	}
	return nil
}
