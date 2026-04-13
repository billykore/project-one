package domain

import (
	"errors"
	"net/mail"
	"time"
)

// User is the core domain entity representing a user in the system.
type User struct {
	ID        int
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate performs domain-level validation on the User entity.
func (u *User) Validate() error {
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
