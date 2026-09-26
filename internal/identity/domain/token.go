package domain

import "time"

// UserToken represents a user session token.
type UserToken struct {
	ID     int
	UserID int
	// Username is transient display data returned after login; session
	// persistence and JWT claims use UserID only.
	Username  string
	Token     string
	ExpiresAt time.Time
}
