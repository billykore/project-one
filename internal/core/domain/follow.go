package domain

import "time"

// Follow represents a follower-followed relationship between users.
type Follow struct {
	FollowerUsername string
	FollowedUsername string
	CreatedAt        time.Time
}

// Following represents a user being followed by the current user with metadata.
type Following struct {
	Username   string
	FirstName  string
	LastName   string
	FollowedAt time.Time
	IsMutual   bool
}

// Follower represents a user following the current user with metadata.
type Follower struct {
	Username   string
	FirstName  string
	LastName   string
	FollowedAt time.Time
	IsMutual   bool
}
