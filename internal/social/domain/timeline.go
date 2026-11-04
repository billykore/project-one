package domain

import "time"

// TimelinePost is a published post as the social feed sees it. It is this
// context's own read projection of a publishing post and carries only the
// fields the feed renders and paginates on.
type TimelinePost struct {
	ID        int
	Author    string
	Title     string
	Content   string
	Tags      []string
	LikeCount int
	CreatedAt time.Time
	UpdatedAt time.Time
}
