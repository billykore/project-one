// Package domain defines core business entities and sentinel error codes.
package domain

import "errors"

// Error codes are general, HTTP-level machine-readable identifiers for API error responses.
// Stable across releases — safe for programmatic dispatch by API consumers.
const (
	CodeNotFound         = "NOT_FOUND"
	CodeInvalidArgument  = "INVALID_ARGUMENT"
	CodeUnauthenticated  = "UNAUTHENTICATED"
	CodePermissionDenied = "PERMISSION_DENIED"
	CodeAlreadyExists    = "ALREADY_EXISTS"
	CodeInternal         = "INTERNAL"
)

var (
	// ErrRepositoryFailure is a sentinel error indicating that a repository operation failed.
	ErrRepositoryFailure = errors.New("repository operation failed")
	// ErrInvalidCredentials is returned when authentication fails due to invalid credentials.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrTokenGenerationFailed is returned when token generation fails.
	ErrTokenGenerationFailed = errors.New("token generation failed")
	// ErrUserNotFound is returned when a user cannot be found in the system.
	ErrUserNotFound = errors.New("user not found")
	// ErrNotificationNotFound is returned when a notification cannot be found in the system.
	ErrNotificationNotFound = errors.New("notification not found")
	// ErrInvalidNotification is returned when a notification is invalid.
	ErrInvalidNotification = errors.New("invalid notification")
	// ErrPostNotFound is returned when a post cannot be found in the system.
	ErrPostNotFound = errors.New("post not found")
	// ErrInvalidPost is returned when post data is invalid.
	ErrInvalidPost = errors.New("invalid post data")
	// ErrEmailAlreadyRegistered is returned when attempting to register an email that is already in use.
	ErrEmailAlreadyRegistered = errors.New("email is already registered")
	// ErrAlreadyFollowing is returned when a user tries to follow someone they already follow.
	ErrAlreadyFollowing = errors.New("already following this user")
	// ErrCannotFollowSelf is returned when a user tries to follow themselves.
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
	// ErrNotFollowing is returned when a user tries to unfollow someone they are not following.
	ErrNotFollowing = errors.New("not following this user")
	// ErrCannotUnfollowSelf is returned when a user tries to unfollow themselves.
	ErrCannotUnfollowSelf = errors.New("cannot unfollow yourself")
	// ErrUsernameAlreadyTaken is returned when attempting to register a username that is already in use.
	ErrUsernameAlreadyTaken = errors.New("username is already taken")
	// ErrInvalidUsername is returned when a username does not meet validation criteria.
	ErrInvalidUsername = errors.New("invalid username")
	// ErrInvalidUser is returned when user data does not meet validation criteria.
	ErrInvalidUser = errors.New("invalid user data")
	// ErrInvalidPassword is returned when a password does not meet validation criteria.
	ErrInvalidPassword = errors.New("invalid password")
	// ErrAlreadyLiked is returned when a user tries to like a post they have already liked.
	ErrCommentNotFound = errors.New("comment not found")
	// ErrCommentNotOwned is returned when a user tries to edit or delete a comment they do not own.
	ErrCommentNotOwned = errors.New("comment not owned by user")
	// ErrInvalidComment is returned when a comment does not meet validation criteria.
	ErrInvalidComment = errors.New("invalid comment")
	// ErrUntrustedToken is returned when a token cannot be trusted (e.g., invalid signature).
	ErrUntrustedToken = errors.New("untrusted token")
	// ErrNotificationNotOwned is returned when a user tries to access or modify a notification they do not own.
	ErrNotificationNotOwned = errors.New("notification not owned by user")
	// ErrPasswordTooShort is returned when a password does not meet the minimum length requirement.
	ErrPasswordTooShort = errors.New("password too short")
	// ErrCommentTooShort is returned when a comment does not meet the minimum length requirement.
	ErrCommentTooShort = errors.New("comment must be at least 1 character")
)
