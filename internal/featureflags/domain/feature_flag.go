package domain

import "time"

// LifecycleState is the lifecycle of a feature flag.
type LifecycleState string

const (
	// LifecycleActive means the flag can be evaluated and edited.
	LifecycleActive LifecycleState = "active"
	// LifecycleArchived means the flag is immutable and always evaluates to its safe default.
	LifecycleArchived LifecycleState = "archived"
)

// Environment is a deployment stage a flag setting applies to.
type Environment string

const (
	EnvironmentLocal      Environment = "local"
	EnvironmentTest       Environment = "test"
	EnvironmentStaging    Environment = "staging"
	EnvironmentProduction Environment = "production"
)

// AvailabilityMode is the per-environment availability strategy for a flag.
type AvailabilityMode string

const (
	// ModeDisabledAll disables the feature for everyone in the environment.
	ModeDisabledAll AvailabilityMode = "disabled_all"
	// ModeEnabledAll enables the feature for everyone in the environment.
	ModeEnabledAll AvailabilityMode = "enabled_all"
	// ModeGradual enables the feature for a percentage of signed-in users.
	ModeGradual AvailabilityMode = "gradual"
)

// OverrideType is the kind of per-user override.
type OverrideType string

const (
	OverrideInclude OverrideType = "include"
	OverrideExclude OverrideType = "exclude"
)

// Decision sources are stable identifiers describing why an evaluation returned its result.
const (
	SourceArchived    = "archived"
	SourceDisabledAll = "disabled_all"
	SourceExclude     = "exclude"
	SourceInclude     = "include"
	SourceGradual     = "gradual"
	SourceEnabledAll  = "enabled_all"
	SourceSafeDefault = "safe_default"
	SourceUnknown     = "unknown"
)

// FeatureFlag is a release-control definition identified by a unique key.
type FeatureFlag struct {
	ID          int
	Key         string
	Name        string
	Purpose     string
	Owner       string
	Lifecycle   LifecycleState
	SafeDefault bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EnvironmentSetting is the state of one flag in one environment.
type EnvironmentSetting struct {
	ID                int
	FlagID            int
	Environment       Environment
	Mode              AvailabilityMode
	RolloutPercentage int
	Revision          int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// UserOverride is an explicit inclusion or exclusion of a signed-in user.
type UserOverride struct {
	ID          int
	FlagID      int
	Environment Environment
	Username    string
	Type        OverrideType
	CreatedAt   time.Time
}

// AuditRecord is an immutable account of an administrative change.
type AuditRecord struct {
	ID            int
	FlagID        int
	Environment   *Environment
	Field         string
	PreviousValue string
	NewValue      string
	Actor         string
	Reason        string
	CreatedAt     time.Time
}

// FeatureFlagDecision is the result of evaluating a flag.
type FeatureFlagDecision struct {
	Key     string
	Enabled bool
	Source  string
	Reason  string
}

// FlagSnapshot is the environment-scoped evaluation data for one flag.
type FlagSnapshot struct {
	Key         string
	Lifecycle   LifecycleState
	SafeDefault bool
	Setting     *EnvironmentSetting
	Overrides   []UserOverride
}

// FeatureFlagDetail bundles a flag with its settings and overrides for review.
type FeatureFlagDetail struct {
	Flag      FeatureFlag
	Settings  []EnvironmentSetting
	Overrides []UserOverride
}
