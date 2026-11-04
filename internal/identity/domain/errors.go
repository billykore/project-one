package domain

import "errors"

// Sentinel errors owned by the identity context.
var (
	// ErrInvalidCredentials is returned when authentication fails due to invalid credentials.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrEmailAlreadyRegistered is returned when attempting to register an email that is already in use.
	ErrEmailAlreadyRegistered = errors.New("email is already registered")
	// ErrUsernameAlreadyTaken is returned when attempting to register a username that is already in use.
	ErrUsernameAlreadyTaken = errors.New("username is already taken")
	// ErrUntrustedToken is returned when a token cannot be trusted (e.g., invalid signature).
	ErrUntrustedToken = errors.New("untrusted token")
	// ErrPasswordTooShort is returned when a password does not meet the minimum length requirement.
	ErrPasswordTooShort = errors.New("password too short")
	// ErrSearchQueryTooShort is returned when a search query doesn't meet the minimum length (3 characters).
	ErrSearchQueryTooShort = errors.New("search query too short")
)
