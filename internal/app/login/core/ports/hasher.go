package ports

import "context"

// Hasher is a driven port for password hashing and verification.
type Hasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Compare(ctx context.Context, password, hash string) error
}
