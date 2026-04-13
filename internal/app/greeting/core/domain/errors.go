package domain

import "errors"

// Sentinel domain errors used across the application.
var (
	ErrGreetingNotFound = errors.New("greeting not found")
)
