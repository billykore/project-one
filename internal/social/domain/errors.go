package domain

import "errors"

// Sentinel errors owned by the social context.
var (
	// ErrAlreadyFollowing is returned when a user tries to follow someone they already follow.
	ErrAlreadyFollowing = errors.New("already following this user")
	// ErrCannotFollowSelf is returned when a user tries to follow themselves.
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
	// ErrNotFollowing is returned when a user tries to unfollow someone they are not following.
	ErrNotFollowing = errors.New("not following this user")
	// ErrCannotUnfollowSelf is returned when a user tries to unfollow themselves.
	ErrCannotUnfollowSelf = errors.New("cannot unfollow yourself")
)
