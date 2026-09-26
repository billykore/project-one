// Package problem defines transport-neutral application error codes and sentinels.
//
// Contexts use this small shared vocabulary only for errors that cross their
// boundaries (for example, repository failures). Context-specific types and
// invariants remain in their owning context.
package problem

import "errors"

// Error codes are general, HTTP-level machine-readable identifiers for API error responses.
// Stable across releases — safe for programmatic dispatch by API consumers.
const (
	CodeNotFound         = "NOT_FOUND"
	CodeInvalidArgument  = "INVALID_ARGUMENT"
	CodeUnauthenticated  = "UNAUTHENTICATED"
	CodePermissionDenied = "PERMISSION_DENIED"
	CodeAlreadyExists    = "ALREADY_EXISTS"
	CodeConflict         = "CONFLICT"
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
	// ErrPostNotOwned is returned when a user tries to modify a post they do not own.
	ErrPostNotOwned = errors.New("post not owned by user")
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
	// ErrPostIDMustBeANumber is returned when a post ID provided in a request is not a valid number.
	ErrPostIDMustBeANumber = errors.New("post ID must be a number")
	// ErrInvalidPostID is returned when a post ID provided in a request is invalid (e.g., negative or zero).
	ErrInvalidPostID = errors.New("invalid post ID")
	// ErrSearchQueryTooShort is returned when a search query doesn't meet the minimum length (3 characters).
	ErrSearchQueryTooShort = errors.New("search query too short")
	// ErrInvalidCursor is returned when a cursor provided in a request is invalid.
	ErrInvalidCursor = errors.New("invalid cursor")
	// ErrFlagNotFound is returned when a feature flag cannot be found.
	ErrFlagNotFound = errors.New("feature flag not found")
	// ErrFlagKeyExists is returned when creating a flag with a duplicate key.
	ErrFlagKeyExists = errors.New("feature flag key already exists")
	// ErrFlagArchived is returned when attempting to edit an archived flag.
	ErrFlagArchived = errors.New("feature flag is archived")
	// ErrRevisionConflict is returned when a flag update conflicts with a newer revision.
	ErrRevisionConflict = errors.New("feature flag revision conflict")
	// ErrInvalidFlagMode is returned when an availability mode is invalid.
	ErrInvalidFlagMode = errors.New("invalid feature flag mode")
	// ErrFlagSettingNotFound is returned when an environment setting does not exist.
	ErrFlagSettingNotFound = errors.New("feature flag setting not found")
	// ErrConflictingOverrides is returned when a user is both included and excluded.
	ErrConflictingOverrides = errors.New("user cannot be both included and excluded")
	// ErrInvalidArgument is returned when feature-flag input is invalid.
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrOperatorOnly is returned when a non-operator attempts to administer flags.
	ErrOperatorOnly = errors.New("operator access required")
	// ErrFeatureDisabled is returned when a guarded action is disabled.
	ErrFeatureDisabled = errors.New("feature is disabled")
)
