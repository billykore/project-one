package domain

import (
	"fmt"
	"net/mail"
	"time"
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
