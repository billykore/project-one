package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/billykore/project-one/internal/app/greeting/core/domain"
	"github.com/billykore/project-one/internal/app/greeting/core/ports"
)

type greetingService struct {
	log  *slog.Logger
	repo ports.GreetingRepository
}

// NewGreetingService creates a new instance of GreetingService.
func NewGreetingService(
	log *slog.Logger,
	repo ports.GreetingRepository,
) ports.GreetingService {
	return &greetingService{
		log:  log.With("layer", "service", "service", "greeting"),
		repo: repo,
	}
}

func (s *greetingService) GetGreeting(ctx context.Context) (*domain.Greeting, error) {
	greeting, err := s.repo.GetGreeting(ctx)
	if err != nil {
		s.log.Error("failed to get greeting from repository", "error", err)
		return nil, fmt.Errorf("get greeting: %w", err)
	}

	if err := greeting.Validate(); err != nil {
		s.log.Warn("invalid greeting found in repository", "error", err)
		return nil, fmt.Errorf("invalid greeting: %w", err)
	}

	s.log.Info("greeting retrieved successfully")
	return greeting, nil
}
