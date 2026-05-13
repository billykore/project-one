package domain

import "time"

// Follow represents a follower-followed relationship between users.
type Follow struct {
	FollowerID int
	FollowedID int
	CreatedAt  time.Time
}

// Following represents a user being followed by the current user with metadata.
type Following struct {
	ID         int
	FirstName  string
	LastName   string
	FollowedAt time.Time
	IsMutual   bool
}
