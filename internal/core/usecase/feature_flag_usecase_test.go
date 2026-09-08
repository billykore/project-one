package usecase

import (
	"context"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	"go.uber.org/mock/gomock"
)

type featureFlagUseCaseLogger struct{}

func (featureFlagUseCaseLogger) Debug(context.Context, string, ...any) {}
func (featureFlagUseCaseLogger) Info(context.Context, string, ...any)  {}
func (featureFlagUseCaseLogger) Warn(context.Context, string, ...any)  {}
func (featureFlagUseCaseLogger) Error(context.Context, string, ...any) {}
func (featureFlagUseCaseLogger) Fatal(context.Context, string, ...any) {}

type featureFlagUseCaseEvaluator struct{}

func (featureFlagUseCaseEvaluator) Evaluate(context.Context, string, string) domain.FeatureFlagDecision {
	return domain.FeatureFlagDecision{Enabled: true}
}
func (featureFlagUseCaseEvaluator) Refresh(context.Context) error { return nil }

func TestFeatureFlagUseCaseCreateFlag(t *testing.T) {
	controller := gomock.NewController(t)
	repository := mocks.NewMockFeatureFlagRepository(controller)
	repository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, flag *domain.FeatureFlag) error {
		flag.ID = 7
		return nil
	})
	repository.EXPECT().AppendAudit(gomock.Any(), gomock.Any()).Return(nil)

	useCase := NewFeatureFlagUseCase(repository, featureFlagUseCaseEvaluator{}, featureFlagUseCaseLogger{})
	flag, err := useCase.CreateFlag(context.Background(), "operator", " Beta.Editor ", "Beta editor", "Try the editor", "content", false)
	if err != nil {
		t.Fatal(err)
	}
	if flag.Key != "beta.editor" || flag.ID != 7 || flag.Lifecycle != domain.LifecycleActive {
		t.Fatalf("unexpected flag: %+v", flag)
	}
}

func TestFeatureFlagUseCaseRejectsInvalidMetadata(t *testing.T) {
	controller := gomock.NewController(t)
	repository := mocks.NewMockFeatureFlagRepository(controller)
	useCase := NewFeatureFlagUseCase(repository, featureFlagUseCaseEvaluator{}, featureFlagUseCaseLogger{})

	if _, err := useCase.CreateFlag(context.Background(), "operator", "", "", "", "", false); err != domain.ErrInvalidArgument {
		t.Fatalf("got %v, want %v", err, domain.ErrInvalidArgument)
	}
}

func TestFeatureFlagUseCaseSupportsGradualRolloutAndRejectsConflicts(t *testing.T) {
	controller := gomock.NewController(t)
	repository := mocks.NewMockFeatureFlagRepository(controller)
	flag := &domain.FeatureFlag{ID: 8, Key: "editor", Lifecycle: domain.LifecycleActive}
	repository.EXPECT().GetByKey(gomock.Any(), "editor").Return(flag, nil)
	repository.EXPECT().GetSetting(gomock.Any(), 8, domain.EnvironmentProduction).Return(nil, domain.ErrFlagSettingNotFound)
	repository.EXPECT().UpsertSetting(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, setting *domain.EnvironmentSetting) error {
		setting.ID = 3
		setting.Revision = 1
		return nil
	})
	repository.EXPECT().AppendAudit(gomock.Any(), gomock.Any()).Return(nil)

	useCase := NewFeatureFlagUseCase(repository, featureFlagUseCaseEvaluator{}, featureFlagUseCaseLogger{})
	setting, err := useCase.SetEnvironment(context.Background(), "operator", "EDITOR", domain.EnvironmentProduction, domain.ModeGradual, 25, 0, "pilot rollout")
	if err != nil || setting.Mode != domain.ModeGradual || setting.RolloutPercentage != 25 {
		t.Fatalf("unexpected gradual setting: %+v, error: %v", setting, err)
	}

	if err := useCase.SetOverrides(context.Background(), "operator", "editor", domain.EnvironmentProduction, []string{"alice"}, []string{"alice"}, "conflict"); err != domain.ErrConflictingOverrides {
		t.Fatalf("got %v, want %v", err, domain.ErrConflictingOverrides)
	}
}

func TestFeatureFlagUseCaseArchivesFlag(t *testing.T) {
	controller := gomock.NewController(t)
	repository := mocks.NewMockFeatureFlagRepository(controller)
	repository.EXPECT().GetByKey(gomock.Any(), "editor").Return(&domain.FeatureFlag{ID: 9, Key: "editor", Lifecycle: domain.LifecycleActive}, nil)
	repository.EXPECT().Archive(gomock.Any(), "editor").Return(nil)
	repository.EXPECT().AppendAudit(gomock.Any(), gomock.Any()).Return(nil)

	useCase := NewFeatureFlagUseCase(repository, featureFlagUseCaseEvaluator{}, featureFlagUseCaseLogger{})
	if err := useCase.Archive(context.Background(), "operator", "editor", "legacy path removed"); err != nil {
		t.Fatal(err)
	}
}
