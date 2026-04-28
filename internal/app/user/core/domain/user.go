package domain

import (
	"errors"
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
		return errors.New("first name is required")
	}
	if len(u.FirstName) < 3 {
		return errors.New("first name must be at least 3 characters")
	}
	if u.LastName == "" {
		return errors.New("last name is required")
	}
	if len(u.LastName) < 3 {
		return errors.New("last name must be at least 3 characters")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return errors.New("invalid email format")
	}
	if len(u.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}
