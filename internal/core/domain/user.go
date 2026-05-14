package domain

import (
	"errors"
	"fmt"
	"net/mail"
	"time"
)

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
	// ErrAlreadyFollowing is returned when a user tries to follow someone they already follow.
	ErrAlreadyFollowing = errors.New("already following this user")
	// ErrCannotFollowSelf is returned when a user tries to follow themselves.
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
	// ErrNotFollowing is returned when a user tries to unfollow someone they are not following.
	ErrNotFollowing = errors.New("not following this user")
	// ErrCannotUnfollowSelf is returned when a user tries to unfollow themselves.
	ErrCannotUnfollowSelf = errors.New("cannot unfollow yourself")
)

// User is the core domain entity representing a user in the system.
type User struct {
	// ID is the unique identifier for the user.
	ID int
	// Email is the user's primary email address.
	Email string
	// Password is the hashed password of the user.
	Password string
	// FirstName is the user's first name.
	FirstName string
	// LastName is the user's last name.
	LastName string
	// CreatedAt is the timestamp when the user was created.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when the user was last updated.
	UpdatedAt time.Time
}

// Validate performs domain-level validation on the User entity.
func (u *User) Validate() error {
	if u.FirstName == "" {
		return fmt.Errorf("%w: first name is required", ErrValidationFailed)
	}
	if len(u.FirstName) < 3 {
		return fmt.Errorf("%w: first name must be at least 3 characters", ErrValidationFailed)
	}
	if u.LastName == "" {
		return fmt.Errorf("%w: last name is required", ErrValidationFailed)
	}
	if len(u.LastName) < 3 {
		return fmt.Errorf("%w: last name must be at least 3 characters", ErrValidationFailed)
	}
	if u.Email == "" {
		return fmt.Errorf("%w: email is required", ErrValidationFailed)
	}
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return fmt.Errorf("%w: invalid email format", ErrValidationFailed)
	}
	if len(u.Password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", ErrValidationFailed)
	}
	return nil
}
