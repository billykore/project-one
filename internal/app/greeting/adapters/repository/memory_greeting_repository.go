package repository

import (
	"context"

	"github.com/billykore/project-one/internal/app/greeting/core/domain"
	"github.com/billykore/project-one/internal/app/greeting/core/ports"
)

type memoryGreetingRepository struct {
	message string
}

// NewMemoryGreetingRepository creates a new instance of GreetingRepository.
func NewMemoryGreetingRepository() ports.GreetingRepository {
	return &memoryGreetingRepository{
		message: "Hello from the dependency-free template!",
	}
}

func (r *memoryGreetingRepository) GetGreeting(ctx context.Context) (*domain.Greeting, error) {
	return &domain.Greeting{
		Message: r.message,
	}, nil
}
