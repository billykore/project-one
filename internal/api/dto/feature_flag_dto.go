package dto

import "time"

// CreateFeatureFlagRequest is the request body for creating a flag.
type CreateFeatureFlagRequest struct {
	Key         string `json:"key" validate:"required,max=128"`
	Name        string `json:"name" validate:"required,max=255"`
	Purpose     string `json:"purpose" validate:"required"`
	Owner       string `json:"owner" validate:"required,max=255"`
	SafeDefault bool   `json:"safeDefault"`
}

// UpdateFeatureFlagRequest is the request body for updating flag metadata.
type UpdateFeatureFlagRequest struct {
	Name        string `json:"name" validate:"required,max=255"`
	Purpose     string `json:"purpose" validate:"required"`
	Owner       string `json:"owner" validate:"required,max=255"`
	SafeDefault bool   `json:"safeDefault"`
	Reason      string `json:"reason" validate:"required"`
}

// SetFeatureFlagEnvironmentRequest changes one environment's availability.
type SetFeatureFlagEnvironmentRequest struct {
	Mode              string `json:"mode" validate:"required"`
	RolloutPercentage int    `json:"rolloutPercentage" validate:"min=0,max=100"`
	Revision          int    `json:"revision" validate:"min=0"`
	Reason            string `json:"reason" validate:"required"`
}

// SetFeatureFlagOverridesRequest replaces user overrides for one environment.
type SetFeatureFlagOverridesRequest struct {
	Environment string   `json:"environment" validate:"required"`
	Include     []string `json:"include"`
	Exclude     []string `json:"exclude"`
	Reason      string   `json:"reason" validate:"required"`
}

// FeatureFlagArchiveRequest archives a flag permanently.
type FeatureFlagArchiveRequest struct {
	Reason string `json:"reason" validate:"required"`
}

// FeatureFlagResponse is the public flag representation.
type FeatureFlagResponse struct {
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Purpose     string    `json:"purpose"`
	Owner       string    `json:"owner"`
	Lifecycle   string    `json:"lifecycle"`
	SafeDefault bool      `json:"safeDefault"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// FeatureFlagSettingResponse is the public environment setting representation.
type FeatureFlagSettingResponse struct {
	Environment       string `json:"environment"`
	Mode              string `json:"mode"`
	RolloutPercentage int    `json:"rolloutPercentage"`
	Revision          int    `json:"revision"`
}

// FeatureFlagOverrideResponse is a public user override representation.
type FeatureFlagOverrideResponse struct {
	Environment string `json:"environment"`
	Username    string `json:"username"`
	Type        string `json:"type"`
}

// FeatureFlagDetailResponse contains the current flag configuration.
type FeatureFlagDetailResponse struct {
	FeatureFlagResponse
	Settings  []FeatureFlagSettingResponse  `json:"settings"`
	Overrides []FeatureFlagOverrideResponse `json:"overrides"`
}

// FeatureFlagListResponse contains all feature flags.
type FeatureFlagListResponse struct {
	Flags []FeatureFlagResponse `json:"flags"`
}

// FeatureFlagDecisionResponse is a safe evaluation result.
type FeatureFlagDecisionResponse struct {
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
	Source  string `json:"source"`
}

// FeatureFlagEvaluateResponse contains decisions for requested keys.
type FeatureFlagEvaluateResponse struct {
	Environment string                        `json:"environment"`
	Decisions   []FeatureFlagDecisionResponse `json:"decisions"`
}

// FeatureFlagAuditResponse is a public audit entry.
type FeatureFlagAuditResponse struct {
	Field         string    `json:"field"`
	Environment   string    `json:"environment,omitempty"`
	PreviousValue string    `json:"previousValue"`
	NewValue      string    `json:"newValue"`
	Actor         string    `json:"actor"`
	Reason        string    `json:"reason"`
	CreatedAt     time.Time `json:"createdAt"`
}

// FeatureFlagAuditListResponse contains a page of audit records.
type FeatureFlagAuditListResponse struct {
	Items      []FeatureFlagAuditResponse `json:"items"`
	NextCursor *int                       `json:"nextCursor"`
	HasMore    bool                       `json:"hasMore"`
}
