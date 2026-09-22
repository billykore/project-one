package dto

// LivenessResponse is the non-sensitive body of the liveness probe.
// It never depends on an external dependency.
type LivenessResponse struct {
	// Status is always "ok" while the process can respond.
	Status string `json:"status" example:"ok"`
	// CheckedAt is the RFC 3339 UTC time the probe ran.
	CheckedAt string `json:"checked_at" example:"2026-09-21T12:00:00Z"`
}

// HealthComponentResponse is the non-sensitive state of one required dependency.
// It deliberately excludes addresses, credentials, and raw dependency errors.
type HealthComponentResponse struct {
	// Name is the stable component name, e.g. "database" or "rabbitmq".
	Name string `json:"name" example:"database"`
	// Status is one of up, down, or unknown.
	Status string `json:"status" example:"up"`
	// CheckedAt is the RFC 3339 UTC time this component was assessed.
	CheckedAt string `json:"checked_at" example:"2026-09-21T12:00:00Z"`
}

// HealthReportResponse is the readiness report body.
type HealthReportResponse struct {
	// Status is ready only when every required component is up.
	Status string `json:"status" example:"ready"`
	// CheckedAt is the RFC 3339 UTC time the assessment completed.
	CheckedAt string `json:"checked_at" example:"2026-09-21T12:00:00Z"`
	// Components holds the state of every assessed dependency.
	Components []HealthComponentResponse `json:"components"`
}
