package ports

import (
	"context"

	"github.com/billykore/project-one/internal/operations/domain"
)

// DependencyChecker is a driven port that assesses one required dependency.
type DependencyChecker interface {
	// Name returns the stable, non-sensitive component name, e.g. "database".
	Name() string
	// Check returns nil while the dependency is available and an error otherwise.
	// Implementations must honour the context deadline so an assessment stays bounded.
	Check(ctx context.Context) error
}

// HealthObserver is a driven port notified after every readiness assessment.
type HealthObserver interface {
	// ObserveAssessment records the latest readiness assessment.
	ObserveAssessment(ctx context.Context, assessment domain.HealthAssessment)
}

// HTTPMetrics is a driven port that records bounded HTTP request observations.
// Implementations must only ever see bounded, non-user-specific label values.
type HTTPMetrics interface {
	// ObserveRequest records one completed request.
	ObserveRequest(ctx context.Context, method, route, statusClass string)
	// ObserveRequestDuration records the duration, in seconds, of one completed request.
	ObserveRequestDuration(ctx context.Context, method, route string, seconds float64)
}

// HealthAssessor is a driving port that produces the current health assessment.
type HealthAssessor interface {
	// Assess evaluates every configured dependency and returns the readiness report.
	// It reports failure through component states rather than an error so the
	// report stays obtainable while a dependency is unavailable.
	Assess(ctx context.Context) domain.HealthAssessment
}
