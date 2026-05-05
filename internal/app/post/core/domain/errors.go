package domain

import "errors"

var (
	ErrPostNotFound   = errors.New("post not found")
	ErrInvalidPost    = errors.New("invalid post data")
	ErrInternalServer = errors.New("internal server error")
)
