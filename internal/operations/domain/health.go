package domain

import "time"

// ComponentStatus is the outcome of a single dependency check.
type ComponentStatus string

const (
	// ComponentStatusUp means the dependency answered within the assessment limit.
	ComponentStatusUp ComponentStatus = "up"
	// ComponentStatusDown means the dependency reported a failure.
	ComponentStatusDown ComponentStatus = "down"
	// ComponentStatusUnknown means the dependency could not be assessed, e.g. it timed out.
	ComponentStatusUnknown ComponentStatus = "unknown"
)

// ReadinessStatus is the overall readiness of the application.
type ReadinessStatus string

const (
	// ReadinessStatusReady means every required component is up.
	ReadinessStatusReady ReadinessStatus = "ready"
	// ReadinessStatusNotReady means at least one required component is not up.
	ReadinessStatusNotReady ReadinessStatus = "not_ready"
)

// ComponentHealth is the non-sensitive state of one required dependency.
// It intentionally carries no addresses, credentials, or raw dependency errors.
type ComponentHealth struct {
	// Name is the stable component name, e.g. "database" or "rabbitmq".
	Name string
	// Status is the result of the most recent check.
	Status ComponentStatus
	// CheckedAt is the UTC time this component was assessed.
	CheckedAt time.Time
}

// HealthAssessment is a timestamped determination of readiness.
type HealthAssessment struct {
	// Status is the overall readiness of the application.
	Status ReadinessStatus
	// CheckedAt is the UTC time the assessment completed.
	CheckedAt time.Time
	// Components holds the state of every assessed dependency, in checker order.
	Components []ComponentHealth
}

// Ready reports whether the application can serve normal user traffic.
// Components that could not be assessed are not treated as up.
func (a HealthAssessment) Ready() bool {
	return a.Status == ReadinessStatusReady
}
