package featureflag

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
)

type evaluatorRepository struct {
	snapshots []domain.FlagSnapshot
	err       error
}

func (r *evaluatorRepository) Create(context.Context, *domain.FeatureFlag) error { return nil }
func (r *evaluatorRepository) GetByKey(context.Context, string) (*domain.FeatureFlag, error) {
	return nil, domain.ErrFlagNotFound
}
func (r *evaluatorRepository) List(context.Context) ([]domain.FeatureFlag, error)        { return nil, nil }
func (r *evaluatorRepository) UpdateMetadata(context.Context, *domain.FeatureFlag) error { return nil }
func (r *evaluatorRepository) Archive(context.Context, string) error                     { return nil }
func (r *evaluatorRepository) GetSetting(context.Context, int, domain.Environment) (*domain.EnvironmentSetting, error) {
	return nil, domain.ErrFlagSettingNotFound
}
func (r *evaluatorRepository) ListSettings(context.Context, int) ([]domain.EnvironmentSetting, error) {
	return nil, nil
}
func (r *evaluatorRepository) UpsertSetting(context.Context, *domain.EnvironmentSetting) error {
	return nil
}
func (r *evaluatorRepository) ListOverrides(context.Context, int, domain.Environment) ([]domain.UserOverride, error) {
	return nil, nil
}
func (r *evaluatorRepository) SetOverrides(context.Context, int, domain.Environment, []string, []string) error {
	return nil
}
func (r *evaluatorRepository) AppendAudit(context.Context, *domain.AuditRecord) error { return nil }
func (r *evaluatorRepository) ListAudit(context.Context, int, int, int) ([]domain.AuditRecord, bool, error) {
	return nil, false, nil
}
func (r *evaluatorRepository) LoadSnapshot(context.Context, domain.Environment) ([]domain.FlagSnapshot, error) {
	return r.snapshots, r.err
}

type evaluatorLogger struct{}

func (evaluatorLogger) Debug(context.Context, string, ...any) {}
func (evaluatorLogger) Info(context.Context, string, ...any)  {}
func (evaluatorLogger) Warn(context.Context, string, ...any)  {}
func (evaluatorLogger) Error(context.Context, string, ...any) {}
func (evaluatorLogger) Fatal(context.Context, string, ...any) {}

func TestEvaluatorPrecedenceAndSafeDefaults(t *testing.T) {
	repository := &evaluatorRepository{snapshots: []domain.FlagSnapshot{
		{Key: "disabled", SafeDefault: true, Setting: &domain.EnvironmentSetting{Mode: domain.ModeDisabledAll}, Overrides: []domain.UserOverride{{Username: "alice", Type: domain.OverrideInclude}}},
		{Key: "rollout", SafeDefault: false, Setting: &domain.EnvironmentSetting{Mode: domain.ModeGradual, RolloutPercentage: 100}, Overrides: []domain.UserOverride{{Username: "alice", Type: domain.OverrideExclude}}},
		{Key: "enabled", Setting: &domain.EnvironmentSetting{Mode: domain.ModeEnabledAll}},
		{Key: "archived", Lifecycle: domain.LifecycleArchived, SafeDefault: true, Setting: &domain.EnvironmentSetting{Mode: domain.ModeEnabledAll}},
	}}
	evaluator, err := NewEvaluator(repository, evaluatorLogger{}, domain.EnvironmentTest, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		key, username, source string
		enabled               bool
	}{
		{"disabled", "alice", domain.SourceDisabledAll, false},
		{"rollout", "alice", domain.SourceExclude, false},
		{"rollout", "bob", domain.SourceGradual, true},
		{"enabled", "", domain.SourceEnabledAll, true},
		{"archived", "bob", domain.SourceArchived, true},
		{"missing", "bob", domain.SourceUnknown, false},
	}
	for _, check := range checks {
		decision := evaluator.Evaluate(context.Background(), check.key, check.username)
		if decision.Enabled != check.enabled || decision.Source != check.source {
			t.Errorf("%s: got enabled=%t source=%q, want enabled=%t source=%q", check.key, decision.Enabled, decision.Source, check.enabled, check.source)
		}
	}
}

func TestEvaluatorRolloutIsDeterministicAndDistributed(t *testing.T) {
	repository := &evaluatorRepository{snapshots: []domain.FlagSnapshot{{
		Key: "gradual", Setting: &domain.EnvironmentSetting{Mode: domain.ModeGradual, RolloutPercentage: 50},
	}}}
	evaluator, err := NewEvaluator(repository, evaluatorLogger{}, domain.EnvironmentProduction, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	first := evaluator.Evaluate(context.Background(), "gradual", "user-123")
	for range 100 {
		if next := evaluator.Evaluate(context.Background(), "gradual", "user-123"); next.Enabled != first.Enabled {
			t.Fatal("same user received different rollout decisions")
		}
	}

	enabled := 0
	for i := 0; i < 10000; i++ {
		if evaluator.Evaluate(context.Background(), "gradual", "user-"+itoa(i)).Enabled {
			enabled++
		}
	}
	percentage := float64(enabled) / 100
	if percentage < 45 || percentage > 55 {
		t.Fatalf("rollout distribution %.2f%% is outside 5 percentage points", percentage)
	}
}

func TestEvaluatorRefreshReplacesSnapshot(t *testing.T) {
	repository := &evaluatorRepository{snapshots: []domain.FlagSnapshot{{Key: "switch", Setting: &domain.EnvironmentSetting{Mode: domain.ModeDisabledAll}}}}
	evaluator, err := NewEvaluator(repository, evaluatorLogger{}, domain.EnvironmentLocal, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	repository.snapshots = []domain.FlagSnapshot{{Key: "switch", Setting: &domain.EnvironmentSetting{Mode: domain.ModeEnabledAll}}}
	if err := evaluator.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !evaluator.Evaluate(context.Background(), "switch", "user").Enabled {
		t.Fatal("refresh did not publish the new snapshot")
	}
}

func TestEvaluatorRefreshErrorKeepsLastSnapshot(t *testing.T) {
	repository := &evaluatorRepository{snapshots: []domain.FlagSnapshot{{Key: "switch", Setting: &domain.EnvironmentSetting{Mode: domain.ModeEnabledAll}}}}
	evaluator, err := NewEvaluator(repository, evaluatorLogger{}, domain.EnvironmentLocal, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	repository.err = errors.New("database unavailable")
	if err := evaluator.Refresh(context.Background()); err == nil {
		t.Fatal("expected refresh error")
	}
	if !evaluator.Evaluate(context.Background(), "switch", "user").Enabled {
		t.Fatal("failed refresh discarded the last valid snapshot")
	}
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	result := ""
	for value > 0 {
		result = string(rune('0'+value%10)) + result
		value /= 10
	}
	return result
}
