package domain

import "errors"

// Sentinel errors owned by the publishing context.
var (
	// ErrPostNotFound is returned when a post cannot be found in the system.
	ErrPostNotFound = errors.New("post not found")
	// ErrPostNotOwned is returned when a user tries to modify a post they do not own.
	ErrPostNotOwned = errors.New("post not owned by user")
	// ErrInvalidPost is returned when post data is invalid.
	ErrInvalidPost = errors.New("invalid post data")
	// ErrCommentNotFound is returned when a comment cannot be found.
	ErrCommentNotFound = errors.New("comment not found")
	// ErrCommentNotOwned is returned when a user tries to edit or delete a comment they do not own.
	ErrCommentNotOwned = errors.New("comment not owned by user")
	// ErrInvalidComment is returned when a comment does not meet validation criteria.
	ErrInvalidComment = errors.New("invalid comment")
	// ErrCommentTooShort is returned when a comment does not meet the minimum length requirement.
	ErrCommentTooShort = errors.New("comment must be at least 1 character")
	// ErrPostIDMustBeANumber is returned when a post ID provided in a request is not a valid number.
	ErrPostIDMustBeANumber = errors.New("post ID must be a number")
	// ErrInvalidPostID is returned when a post ID provided in a request is invalid (e.g., negative or zero).
	ErrInvalidPostID = errors.New("invalid post ID")
)
