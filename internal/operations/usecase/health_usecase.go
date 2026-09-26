package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/billykore/project-one/internal/operations/domain"
	"github.com/billykore/project-one/internal/operations/ports"
	platformports "github.com/billykore/project-one/internal/platform/ports"
)

// DefaultHealthCheckTimeout bounds a single dependency check when no explicit
// timeout is configured for the health use case.
const DefaultHealthCheckTimeout = 2 * time.Second

// HealthUseCase aggregates dependency checks into a readiness assessment.
// It knows nothing about HTTP, Prometheus, or concrete dependencies.
type HealthUseCase struct {
	checkers []ports.DependencyChecker
	observer ports.HealthObserver
	logger   platformports.Logger
	timeout  time.Duration
}

// NewHealthUseCase creates a health use case. A non-positive checkTimeout falls
// back to DefaultHealthCheckTimeout. observer may be nil when no consumer of the
// assessment is registered.
func NewHealthUseCase(
	checkers []ports.DependencyChecker,
	observer ports.HealthObserver,
	checkTimeout time.Duration,
	logger platformports.Logger,
) *HealthUseCase {
	if checkTimeout <= 0 {
		checkTimeout = DefaultHealthCheckTimeout
	}
	return &HealthUseCase{
		checkers: checkers,
		observer: observer,
		logger:   logger,
		timeout:  checkTimeout,
	}
}

// Assess evaluates every configured dependency with a bounded context and
// returns the readiness report. Each dependency gets its own deadline, so one
// unresponsive dependency cannot hide the state of the others.
func (u *HealthUseCase) Assess(ctx context.Context) domain.HealthAssessment {
	components := make([]domain.ComponentHealth, 0, len(u.checkers))
	for _, checker := range u.checkers {
		components = append(components, u.checkComponent(ctx, checker))
	}

	status := domain.ReadinessStatusReady
	for _, component := range components {
		if component.Status != domain.ComponentStatusUp {
			status = domain.ReadinessStatusNotReady
			break
		}
	}

	assessment := domain.HealthAssessment{
		Status:     status,
		CheckedAt:  time.Now().UTC(),
		Components: components,
	}

	if u.observer != nil {
		u.observer.ObserveAssessment(ctx, assessment)
	}

	return assessment
}

func (u *HealthUseCase) checkComponent(ctx context.Context, checker ports.DependencyChecker) domain.ComponentHealth {
	checkCtx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	err := checker.Check(checkCtx)

	component := domain.ComponentHealth{
		Name:      checker.Name(),
		Status:    domain.ComponentStatusUp,
		CheckedAt: time.Now().UTC(),
	}
	if err == nil {
		return component
	}

	component.Status = domain.ComponentStatusDown
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(checkCtx.Err(), context.DeadlineExceeded) {
		component.Status = domain.ComponentStatusUnknown
	}

	// Diagnostic detail stays server-side; the report never carries raw errors.
	if u.logger != nil {
		u.logger.Warn(ctx, "health check failed",
			"component", component.Name,
			"status", string(component.Status),
			"error", err,
		)
	}

	return component
}
