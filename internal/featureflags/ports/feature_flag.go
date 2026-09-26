package ports

import (
	"context"

	"github.com/billykore/project-one/internal/featureflags/domain"
	vo "github.com/billykore/project-one/internal/platform/pagination"
)

// FeatureFlagRepository is a driven port for feature-flag persistence.
type FeatureFlagRepository interface {
	// Create saves a new feature flag.
	Create(ctx context.Context, flag *domain.FeatureFlag) error
	// GetByKey retrieves a flag by its unique key.
	GetByKey(ctx context.Context, key string) (*domain.FeatureFlag, error)
	// List retrieves all feature flags ordered by key.
	List(ctx context.Context) ([]domain.FeatureFlag, error)
	// UpdateMetadata updates the mutable metadata of a flag.
	UpdateMetadata(ctx context.Context, flag *domain.FeatureFlag) error
	// Archive marks a flag as archived.
	Archive(ctx context.Context, key string) error

	// GetSetting retrieves the environment setting for a flag, or ErrFlagSettingNotFound.
	GetSetting(ctx context.Context, flagID int, env domain.Environment) (*domain.EnvironmentSetting, error)
	// ListSettings retrieves all environment settings for a flag.
	ListSettings(ctx context.Context, flagID int) ([]domain.EnvironmentSetting, error)
	// UpsertSetting creates or updates a setting using revision-based optimistic concurrency.
	// On revision mismatch it returns problem.ErrRevisionConflict.
	UpsertSetting(ctx context.Context, setting *domain.EnvironmentSetting) error

	// ListOverrides retrieves the overrides for a flag and environment.
	ListOverrides(ctx context.Context, flagID int, env domain.Environment) ([]domain.UserOverride, error)
	// SetOverrides atomically replaces the overrides for a flag and environment.
	SetOverrides(ctx context.Context, flagID int, env domain.Environment, includes, excludes []string) error

	// AppendAudit appends an immutable audit record.
	AppendAudit(ctx context.Context, record *domain.AuditRecord) error
	// ListAudit returns a page of audit records for a flag (newest first).
	ListAudit(ctx context.Context, flagID int, cursor *vo.Cursor, limit int) ([]domain.AuditRecord, bool, error)

	// LoadSnapshot loads the environment-scoped evaluation data for all flags.
	LoadSnapshot(ctx context.Context, env domain.Environment) ([]domain.FlagSnapshot, error)
}

// FeatureFlagEvaluator is a driven port for evaluating feature flags.
type FeatureFlagEvaluator interface {
	// Evaluate returns a decision for the flag and optional signed-in username.
	Evaluate(ctx context.Context, key string, username string) domain.FeatureFlagDecision
	// Refresh reloads flag state from the repository immediately.
	Refresh(ctx context.Context) error
}

// FeatureFlagUseCase is a driving port for feature-flag administration and evaluation.
type FeatureFlagUseCase interface {
	// CreateFlag creates a new feature flag and persists the initial metadata for the actor.
	CreateFlag(ctx context.Context, actor, key, name, purpose, owner string, safeDefault bool) (*domain.FeatureFlag, error)
	// GetFlagDetail returns the full metadata and environment configuration for a flag.
	GetFlagDetail(ctx context.Context, key string) (*domain.FeatureFlagDetail, error)
	// ListFlags returns all feature flags in the configured registry.
	ListFlags(ctx context.Context) ([]domain.FeatureFlag, error)
	// UpdateFlag updates the mutable metadata for an existing flag and records the change as an audit event.
	UpdateFlag(ctx context.Context, actor, key, name, purpose, owner string, safeDefault bool, reason string) (*domain.FeatureFlag, error)
	// SetEnvironment configures a flag's rollout behavior for a specific environment, including mode and percentage.
	SetEnvironment(ctx context.Context, actor, key string, env domain.Environment, mode domain.AvailabilityMode, percentage, revision int, reason string) (*domain.EnvironmentSetting, error)
	// SetOverrides replaces the include and exclude user lists for a flag in a given environment.
	SetOverrides(ctx context.Context, actor, key string, env domain.Environment, includes, excludes []string, reason string) error
	// Archive marks a flag as archived and prevents future evaluation unless explicitly re-enabled by policy.
	Archive(ctx context.Context, actor, key, reason string) error
	// ListAudit returns the audit history for a flag, ordered newest-first with pagination support.
	ListAudit(ctx context.Context, key string, cursor *vo.Cursor, limit int) ([]domain.AuditRecord, bool, error)
	// Evaluate resolves the current decision for a flag for the provided username, respecting overrides and rollout rules.
	Evaluate(ctx context.Context, key string, username string) domain.FeatureFlagDecision
}
