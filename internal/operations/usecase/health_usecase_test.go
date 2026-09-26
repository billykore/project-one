package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/operations/domain"
	"github.com/billykore/project-one/internal/operations/ports"
	"github.com/billykore/project-one/internal/testkit/mocks"
	"go.uber.org/mock/gomock"
)

type healthUseCaseLogger struct{}

func (healthUseCaseLogger) Debug(context.Context, string, ...any) {}
func (healthUseCaseLogger) Info(context.Context, string, ...any)  {}
func (healthUseCaseLogger) Warn(context.Context, string, ...any)  {}
func (healthUseCaseLogger) Error(context.Context, string, ...any) {}
func (healthUseCaseLogger) Fatal(context.Context, string, ...any) {}

// newChecker returns a dependency checker mock with a fixed component name.
func newChecker(controller *gomock.Controller, name string) *mocks.MockDependencyChecker {
	checker := mocks.NewMockDependencyChecker(controller)
	checker.EXPECT().Name().Return(name).AnyTimes()
	return checker
}

func TestHealthUseCaseReportsReadyWhenEveryComponentIsUp(t *testing.T) {
	controller := gomock.NewController(t)
	database := newChecker(controller, "database")
	broker := newChecker(controller, "rabbitmq")
	database.EXPECT().Check(gomock.Any()).Return(nil)
	broker.EXPECT().Check(gomock.Any()).Return(nil)

	before := time.Now().UTC()
	useCase := NewHealthUseCase([]ports.DependencyChecker{database, broker}, nil, time.Second, healthUseCaseLogger{})
	assessment := useCase.Assess(context.Background())
	after := time.Now().UTC()

	if assessment.Status != domain.ReadinessStatusReady || !assessment.Ready() {
		t.Fatalf("status = %q, want %q", assessment.Status, domain.ReadinessStatusReady)
	}
	if len(assessment.Components) != 2 {
		t.Fatalf("component count = %d, want 2", len(assessment.Components))
	}
	if assessment.Components[0].Name != "database" || assessment.Components[1].Name != "rabbitmq" {
		t.Fatalf("component order not preserved: %+v", assessment.Components)
	}
	for _, component := range assessment.Components {
		if component.Status != domain.ComponentStatusUp {
			t.Fatalf("component %q status = %q, want %q", component.Name, component.Status, domain.ComponentStatusUp)
		}
	}
	if assessment.CheckedAt.Location() != time.UTC {
		t.Fatalf("assessment timestamp is not UTC: %v", assessment.CheckedAt.Location())
	}
	if assessment.CheckedAt.Before(before) || assessment.CheckedAt.After(after) {
		t.Fatalf("assessment timestamp %v outside assessment window %v-%v", assessment.CheckedAt, before, after)
	}
}

func TestHealthUseCaseReportsNotReadyAndKeepsEveryComponentState(t *testing.T) {
	controller := gomock.NewController(t)
	database := newChecker(controller, "database")
	broker := newChecker(controller, "rabbitmq")
	database.EXPECT().Check(gomock.Any()).Return(errors.New("dial tcp 10.0.0.5:5432: connection refused"))
	broker.EXPECT().Check(gomock.Any()).Return(nil)

	useCase := NewHealthUseCase([]ports.DependencyChecker{database, broker}, nil, time.Second, healthUseCaseLogger{})
	assessment := useCase.Assess(context.Background())

	if assessment.Ready() || assessment.Status != domain.ReadinessStatusNotReady {
		t.Fatalf("status = %q, want %q", assessment.Status, domain.ReadinessStatusNotReady)
	}
	if len(assessment.Components) != 2 {
		t.Fatalf("component count = %d, want the failed and healthy components preserved", len(assessment.Components))
	}
	if assessment.Components[0].Status != domain.ComponentStatusDown {
		t.Fatalf("failed component status = %q, want %q", assessment.Components[0].Status, domain.ComponentStatusDown)
	}
	if assessment.Components[1].Status != domain.ComponentStatusUp {
		t.Fatalf("healthy component status = %q, want %q", assessment.Components[1].Status, domain.ComponentStatusUp)
	}
}

func TestHealthUseCaseBoundsSlowChecksAndMarksThemUnknown(t *testing.T) {
	controller := gomock.NewController(t)
	database := newChecker(controller, "database")
	database.EXPECT().Check(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	broker := newChecker(controller, "rabbitmq")
	broker.EXPECT().Check(gomock.Any()).Return(nil)

	useCase := NewHealthUseCase([]ports.DependencyChecker{database, broker}, nil, 30*time.Millisecond, healthUseCaseLogger{})

	start := time.Now()
	assessment := useCase.Assess(context.Background())
	elapsed := time.Since(start)

	if elapsed > time.Second {
		t.Fatalf("assessment blocked for %v, want a bounded check", elapsed)
	}
	if assessment.Ready() {
		t.Fatalf("status = %q, want not ready for a timed out component", assessment.Status)
	}
	if assessment.Components[0].Status != domain.ComponentStatusUnknown {
		t.Fatalf("timed out component status = %q, want %q", assessment.Components[0].Status, domain.ComponentStatusUnknown)
	}
	if assessment.Components[1].Status != domain.ComponentStatusUp {
		t.Fatalf("healthy component status = %q, want %q", assessment.Components[1].Status, domain.ComponentStatusUp)
	}
}

func TestHealthUseCaseReturnsAnAssessmentWhenAComponentPanicsFree(t *testing.T) {
	controller := gomock.NewController(t)
	database := newChecker(controller, "database")
	database.EXPECT().Check(gomock.Any()).Return(nil)

	useCase := NewHealthUseCase([]ports.DependencyChecker{database}, nil, 0, healthUseCaseLogger{})
	if assessment := useCase.Assess(context.Background()); !assessment.Ready() {
		t.Fatalf("status = %q, want ready with the default check timeout applied", assessment.Status)
	}
}

func TestHealthUseCaseNotifiesTheObserverWithComponentTimestamps(t *testing.T) {
	controller := gomock.NewController(t)
	database := newChecker(controller, "database")
	database.EXPECT().Check(gomock.Any()).Return(nil)

	var observed domain.HealthAssessment
	observer := mocks.NewMockHealthObserver(controller)
	observer.EXPECT().ObserveAssessment(gomock.Any(), gomock.Any()).Do(func(_ context.Context, assessment domain.HealthAssessment) {
		observed = assessment
	}).Times(1)

	useCase := NewHealthUseCase([]ports.DependencyChecker{database}, observer, time.Second, healthUseCaseLogger{})
	assessment := useCase.Assess(context.Background())

	if observed.Status != assessment.Status || len(observed.Components) != len(assessment.Components) {
		t.Fatalf("observed assessment %+v does not match returned assessment %+v", observed, assessment)
	}
	component := assessment.Components[0]
	if component.CheckedAt.IsZero() || component.CheckedAt.Location() != time.UTC {
		t.Fatalf("component timestamp not set in UTC: %v", component.CheckedAt)
	}
	if component.CheckedAt.After(assessment.CheckedAt) {
		t.Fatalf("component timestamp %v is after the assessment timestamp %v", component.CheckedAt, assessment.CheckedAt)
	}
}
