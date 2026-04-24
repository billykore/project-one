package ports

import "context"

// Hasher is a driven port for password hashing and verification.
type Hasher interface {
	// Hash returns the hashed version of the password.
	Hash(ctx context.Context, password string) (string, error)
	// Compare checks if the password matches the hash.
	Compare(ctx context.Context, password, hash string) error
}
