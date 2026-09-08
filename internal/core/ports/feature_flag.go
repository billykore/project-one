package ports

import (
	"context"

	"github.com/billykore/project-one/internal/core/domain"
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
	// On revision mismatch it returns domain.ErrRevisionConflict.
	UpsertSetting(ctx context.Context, setting *domain.EnvironmentSetting) error

	// ListOverrides retrieves the overrides for a flag and environment.
	ListOverrides(ctx context.Context, flagID int, env domain.Environment) ([]domain.UserOverride, error)
	// SetOverrides atomically replaces the overrides for a flag and environment.
	SetOverrides(ctx context.Context, flagID int, env domain.Environment, includes, excludes []string) error

	// AppendAudit appends an immutable audit record.
	AppendAudit(ctx context.Context, record *domain.AuditRecord) error
	// ListAudit returns a page of audit records for a flag (newest first).
	ListAudit(ctx context.Context, flagID int, cursor int, limit int) ([]domain.AuditRecord, bool, error)

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
	CreateFlag(ctx context.Context, actor, key, name, purpose, owner string, safeDefault bool) (*domain.FeatureFlag, error)
	GetFlagDetail(ctx context.Context, key string) (*domain.FeatureFlagDetail, error)
	ListFlags(ctx context.Context) ([]domain.FeatureFlag, error)
	UpdateFlag(ctx context.Context, actor, key, name, purpose, owner string, safeDefault bool, reason string) (*domain.FeatureFlag, error)
	SetEnvironment(ctx context.Context, actor, key string, env domain.Environment, mode domain.AvailabilityMode, percentage, revision int, reason string) (*domain.EnvironmentSetting, error)
	SetOverrides(ctx context.Context, actor, key string, env domain.Environment, includes, excludes []string, reason string) error
	Archive(ctx context.Context, actor, key, reason string) error
	ListAudit(ctx context.Context, key string, cursor, limit int) ([]domain.AuditRecord, bool, error)
	Evaluate(ctx context.Context, key string, username string) domain.FeatureFlagDecision
}
