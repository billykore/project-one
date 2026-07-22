package domain

import "time"

// User is the core domain entity representing a user in the system.
type User struct {
	// ID is the unique identifier for the user.
	ID int
	// Email is the user's primary email address.
	Email string
	// Username is the user's unique username.
	Username string
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
