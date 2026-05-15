package domain

import "time"

// UserToken represents a user session token.
type UserToken struct {
	ID        int
	UserID    int
	Username  string
	Token     string
	ExpiresAt time.Time
}
