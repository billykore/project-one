package domain

import "errors"

// Sentinel domain errors used across the application.
var (
	// ErrUserNotFound is returned when a user cannot be found in the system.
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidCredentials is returned when authentication fails due to wrong email or password.
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrUnauthorized is returned when a request lacks valid authentication credentials.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrInternalServer is returned for unexpected server-side errors.
	ErrInternalServer = errors.New("internal server error")
	// ErrEmailAlreadyRegistered is returned when attempting to register an email that is already in use.
	ErrEmailAlreadyRegistered = errors.New("email is already registered")
	// ErrValidationFailed is returned when domain validation fails.
	ErrValidationFailed = errors.New("validation failed")
)
