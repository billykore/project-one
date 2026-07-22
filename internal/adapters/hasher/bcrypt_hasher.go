package hasher

import (
	"context"
	"fmt"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

type bcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new instance of Hasher.
func NewBcryptHasher() ports.Hasher {
	return &bcryptHasher{cost: bcrypt.DefaultCost}
}

func (h *bcryptHasher) Hash(_ context.Context, password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	return string(bytes), err
}

func (h *bcryptHasher) Compare(_ context.Context, password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidCredentials, err)
	}
	return nil
}
