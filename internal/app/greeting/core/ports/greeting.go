package ports

import (
	"context"

	"github.com/billykore/project-one/internal/app/greeting/core/domain"
)

// GreetingRepository is a driven port for greeting persistence.
type GreetingRepository interface {
	GetGreeting(ctx context.Context) (*domain.Greeting, error)
}

// GreetingService is a driving port for greeting-related application logic.
type GreetingService interface {
	GetGreeting(ctx context.Context) (*domain.Greeting, error)
}
