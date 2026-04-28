package domain

import "time"

// TokenDetails holds the token string and its expiration time.
type TokenDetails struct {
	Token     string
	ExpiresAt time.Time
}
