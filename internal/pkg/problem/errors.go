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
	// ErrUserNotFound is returned when a user cannot be found in the system.
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidUsername is returned when a username does not meet validation criteria.
	ErrInvalidUsername = errors.New("invalid username")
	// ErrInvalidUser is returned when user data does not meet validation criteria.
	ErrInvalidUser = errors.New("invalid user data")
	// ErrInvalidCursor is returned when a cursor provided in a request is invalid.
	ErrInvalidCursor = errors.New("invalid cursor")
	// ErrInvalidArgument is returned when feature-flag input is invalid.
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrFeatureDisabled is returned when a guarded action is disabled.
	ErrFeatureDisabled = errors.New("feature is disabled")
)
