package domain

import "errors"

// Sentinel domain errors used across the application.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInternalServer     = errors.New("internal server error")
)
