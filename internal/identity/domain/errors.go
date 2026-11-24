package domain

import (
	"errors"
	"fmt"
	"time"
)

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
	// ErrTooManyLoginAttempts is returned when an account exceeds its login attempt limit.
	ErrTooManyLoginAttempts = errors.New("too many login attempts")
)

// LoginThrottledError reports a rejected login attempt and how long the
// caller must wait before the account accepts another attempt. It matches
// ErrTooManyLoginAttempts with errors.Is.
type LoginThrottledError struct {
	RetryAfter time.Duration
}

func (e *LoginThrottledError) Error() string {
	return fmt.Sprintf("%s: retry after %s", ErrTooManyLoginAttempts, e.RetryAfter)
}

// Unwrap exposes the sentinel so the error maps to its problem details.
func (e *LoginThrottledError) Unwrap() error {
	return ErrTooManyLoginAttempts
}
