package domain

import "time"

// Follow represents a follower-followed relationship between users.
type Follow struct {
	FollowerID int
	FollowedID int
	CreatedAt  time.Time
}
