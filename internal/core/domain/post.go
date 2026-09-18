package domain

import (
	"strings"
	"time"
)

// Post is the core domain entity representing a user's post.
type Post struct {
	ID        int
	UserID    int
	Username  string
	Title     string
	Content   string
	Tags      []string
	LikeCount int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Update applies the editable fields of a post.
func (p *Post) Update(title, content string) {
	if title = strings.TrimSpace(title); title != "" {
		p.Title = title
	}
	if content = strings.TrimSpace(content); content != "" {
		p.Content = content
	}
}

// AddLike updates the aggregate's like-count invariant.
func (p *Post) AddLike() { p.LikeCount++ }

// RemoveLike updates the aggregate's like-count invariant without going negative.
func (p *Post) RemoveLike() {
	if p.LikeCount > 0 {
		p.LikeCount--
	}
}
