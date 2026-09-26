package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/billykore/project-one/internal/featureflags/domain"
	"github.com/billykore/project-one/internal/featureflags/ports"
	vo "github.com/billykore/project-one/internal/platform/pagination"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	"github.com/billykore/project-one/internal/platform/problem"
)

type featureFlagUseCase struct {
	repository ports.FeatureFlagRepository
	evaluator  ports.FeatureFlagEvaluator
	logger     platformports.Logger
}

// NewFeatureFlagUseCase creates the feature-flag application service.
func NewFeatureFlagUseCase(
	repository ports.FeatureFlagRepository,
	evaluator ports.FeatureFlagEvaluator,
	logger platformports.Logger,
) ports.FeatureFlagUseCase {
	if repository == nil || evaluator == nil || logger == nil {
		panic("NewFeatureFlagUseCase: dependencies must not be nil")
	}
	return &featureFlagUseCase{repository: repository, evaluator: evaluator, logger: logger}
}

func (uc *featureFlagUseCase) CreateFlag(ctx context.Context, actor, key, name, purpose, owner string, safeDefault bool) (*domain.FeatureFlag, error) {
	key = normalizeFlagKey(key)
	if err := validateFlagMetadata(key, name, purpose, owner); err != nil {
		return nil, err
	}
	flag := &domain.FeatureFlag{
		Key:         key,
		Name:        strings.TrimSpace(name),
		Purpose:     strings.TrimSpace(purpose),
		Owner:       strings.TrimSpace(owner),
		Lifecycle:   domain.LifecycleActive,
		SafeDefault: safeDefault,
	}
	if err := uc.repository.Create(ctx, flag); err != nil {
		return nil, err
	}
	if err := uc.audit(ctx, flag.ID, nil, "create", "", flag.Key, actor, "flag created"); err != nil {
		return nil, err
	}
	uc.refresh(ctx)
	return flag, nil
}

func (uc *featureFlagUseCase) GetFlagDetail(ctx context.Context, key string) (*domain.FeatureFlagDetail, error) {
	flag, err := uc.repository.GetByKey(ctx, normalizeFlagKey(key))
	if err != nil {
		return nil, err
	}
	settings, err := uc.repository.ListSettings(ctx, flag.ID)
	if err != nil {
		return nil, err
	}
	detail := &domain.FeatureFlagDetail{Flag: *flag, Settings: settings, Overrides: []domain.UserOverride{}}
	for _, setting := range settings {
		overrides, overrideErr := uc.repository.ListOverrides(ctx, flag.ID, setting.Environment)
		if overrideErr != nil {
			return nil, overrideErr
		}
		detail.Overrides = append(detail.Overrides, overrides...)
	}
	return detail, nil
}

func (uc *featureFlagUseCase) ListFlags(ctx context.Context) ([]domain.FeatureFlag, error) {
	return uc.repository.List(ctx)
}

func (uc *featureFlagUseCase) UpdateFlag(ctx context.Context, actor, key, name, purpose, owner string, safeDefault bool, reason string) (*domain.FeatureFlag, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, problem.ErrInvalidArgument
	}
	flag, err := uc.repository.GetByKey(ctx, normalizeFlagKey(key))
	if err != nil {
		return nil, err
	}
	if flag.Lifecycle == domain.LifecycleArchived {
		return nil, problem.ErrFlagArchived
	}
	if err := validateFlagMetadata(flag.Key, name, purpose, owner); err != nil {
		return nil, err
	}
	previous := fmt.Sprintf("name=%s;purpose=%s;owner=%s;safe_default=%t", flag.Name, flag.Purpose, flag.Owner, flag.SafeDefault)
	flag.Name = strings.TrimSpace(name)
	flag.Purpose = strings.TrimSpace(purpose)
	flag.Owner = strings.TrimSpace(owner)
	flag.SafeDefault = safeDefault
	if err := uc.repository.UpdateMetadata(ctx, flag); err != nil {
		return nil, err
	}
	current := fmt.Sprintf("name=%s;purpose=%s;owner=%s;safe_default=%t", flag.Name, flag.Purpose, flag.Owner, flag.SafeDefault)
	if err := uc.audit(ctx, flag.ID, nil, "metadata", previous, current, actor, reason); err != nil {
		return nil, err
	}
	uc.refresh(ctx)
	return flag, nil
}

func (uc *featureFlagUseCase) SetEnvironment(ctx context.Context, actor, key string, env domain.Environment, mode domain.AvailabilityMode, percentage, revision int, reason string) (*domain.EnvironmentSetting, error) {
	if !validEnvironment(env) || !validMode(mode) || percentage < 0 || percentage > 100 || strings.TrimSpace(reason) == "" {
		return nil, problem.ErrInvalidArgument
	}
	if mode != domain.ModeGradual && percentage != 0 && percentage != 100 {
		return nil, problem.ErrInvalidArgument
	}
	flag, err := uc.repository.GetByKey(ctx, normalizeFlagKey(key))
	if err != nil {
		return nil, err
	}
	if flag.Lifecycle == domain.LifecycleArchived {
		return nil, problem.ErrFlagArchived
	}
	previous, getErr := uc.repository.GetSetting(ctx, flag.ID, env)
	if getErr != nil && !errors.Is(getErr, problem.ErrFlagSettingNotFound) {
		return nil, getErr
	}
	setting := &domain.EnvironmentSetting{FlagID: flag.ID, Environment: env, Mode: mode, RolloutPercentage: percentage}
	if previous != nil {
		setting.ID = previous.ID
		setting.Revision = revision
	}
	if err := uc.repository.UpsertSetting(ctx, setting); err != nil {
		return nil, err
	}
	previousValue := ""
	if previous != nil {
		previousValue = fmt.Sprintf("mode=%s;rollout=%d;revision=%d", previous.Mode, previous.RolloutPercentage, previous.Revision)
	}
	currentValue := fmt.Sprintf("mode=%s;rollout=%d;revision=%d", mode, percentage, setting.Revision)
	if err := uc.audit(ctx, flag.ID, &env, "environment_setting", previousValue, currentValue, actor, reason); err != nil {
		return nil, err
	}
	uc.refresh(ctx)
	return setting, nil
}

func (uc *featureFlagUseCase) SetOverrides(ctx context.Context, actor, key string, env domain.Environment, includes, excludes []string, reason string) error {
	if !validEnvironment(env) || strings.TrimSpace(reason) == "" {
		return problem.ErrInvalidArgument
	}
	includeSet := make(map[string]struct{}, len(includes))
	for _, username := range includes {
		includeSet[strings.TrimSpace(username)] = struct{}{}
	}
	for _, username := range excludes {
		if _, exists := includeSet[strings.TrimSpace(username)]; exists {
			return problem.ErrConflictingOverrides
		}
	}
	flag, err := uc.repository.GetByKey(ctx, normalizeFlagKey(key))
	if err != nil {
		return err
	}
	if flag.Lifecycle == domain.LifecycleArchived {
		return problem.ErrFlagArchived
	}
	if err := uc.repository.SetOverrides(ctx, flag.ID, env, normalizedUsernames(includes), normalizedUsernames(excludes)); err != nil {
		return err
	}
	if err := uc.audit(ctx, flag.ID, &env, "overrides", "", fmt.Sprintf("include=%v;exclude=%v", includes, excludes), actor, reason); err != nil {
		return err
	}
	uc.refresh(ctx)
	return nil
}

func (uc *featureFlagUseCase) Archive(ctx context.Context, actor, key, reason string) error {
	if strings.TrimSpace(reason) == "" {
		return problem.ErrInvalidArgument
	}
	flag, err := uc.repository.GetByKey(ctx, normalizeFlagKey(key))
	if err != nil {
		return err
	}
	if flag.Lifecycle == domain.LifecycleArchived {
		return problem.ErrFlagArchived
	}
	if err := uc.repository.Archive(ctx, flag.Key); err != nil {
		return err
	}
	if err := uc.audit(ctx, flag.ID, nil, "lifecycle", string(domain.LifecycleActive), string(domain.LifecycleArchived), actor, reason); err != nil {
		return err
	}
	uc.refresh(ctx)
	return nil
}

func (uc *featureFlagUseCase) ListAudit(ctx context.Context, key string, cursor *vo.Cursor, limit int) ([]domain.AuditRecord, bool, error) {
	flag, err := uc.repository.GetByKey(ctx, normalizeFlagKey(key))
	if err != nil {
		return nil, false, err
	}
	return uc.repository.ListAudit(ctx, flag.ID, cursor, limit)
}

func (uc *featureFlagUseCase) Evaluate(ctx context.Context, key, username string) domain.FeatureFlagDecision {
	return uc.evaluator.Evaluate(ctx, key, username)
}

func (uc *featureFlagUseCase) audit(ctx context.Context, flagID int, env *domain.Environment, field, previous, current, actor, reason string) error {
	return uc.repository.AppendAudit(ctx, &domain.AuditRecord{
		FlagID:        flagID,
		Environment:   env,
		Field:         field,
		PreviousValue: previous,
		NewValue:      current,
		Actor:         actor,
		Reason:        reason,
	})
}

func (uc *featureFlagUseCase) refresh(ctx context.Context) {
	if err := uc.evaluator.Refresh(ctx); err != nil {
		uc.logger.Warn(ctx, "feature flag cache refresh failed after update", "failure_category", "post_write_refresh", "error", err)
	}
}

// normalizeFlagKey trims whitespace and converts the key to lowercase for consistent storage and retrieval.
func normalizeFlagKey(key string) string { return strings.ToLower(strings.TrimSpace(key)) }

// normalizedUsernames trims whitespace from each username and filters out any empty strings.
func normalizedUsernames(usernames []string) []string {
	result := make([]string, 0, len(usernames))
	for _, username := range usernames {
		trimmed := strings.TrimSpace(username)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// validateFlagMetadata checks that the provided metadata for a feature flag is valid.
func validateFlagMetadata(key, name, purpose, owner string) error {
	if key == "" || len(key) > 128 || strings.TrimSpace(name) == "" || strings.TrimSpace(purpose) == "" || strings.TrimSpace(owner) == "" {
		return problem.ErrInvalidArgument
	}
	return nil
}

// validEnvironment checks if the provided environment is one of the recognized environments.
func validEnvironment(environment domain.Environment) bool {
	switch environment {
	case domain.EnvironmentLocal, domain.EnvironmentTest, domain.EnvironmentStaging, domain.EnvironmentProduction:
		return true
	default:
		return false
	}
}

// validMode checks if the provided availability mode is one of the recognized modes.
func validMode(mode domain.AvailabilityMode) bool {
	switch mode {
	case domain.ModeDisabledAll, domain.ModeEnabledAll, domain.ModeGradual:
		return true
	default:
		return false
	}
}
