package health

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/testkit/mocks"
	"go.uber.org/mock/gomock"
)

type stubPinger struct{ err error }

func (s stubPinger) PingContext(context.Context) error { return s.err }

type slowPinger struct{}

func (slowPinger) PingContext(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

type stubHealthReporter struct{ healthy bool }

func (s stubHealthReporter) Healthy() bool { return s.healthy }

func TestDatabaseCheckerReportsUp(t *testing.T) {
	checker := NewDatabaseChecker(stubPinger{})

	if checker.Name() != "database" {
		t.Fatalf("name = %q, want %q", checker.Name(), "database")
	}
	if err := checker.Check(context.Background()); err != nil {
		t.Fatalf("check returned %v, want nil", err)
	}
}

func TestDatabaseCheckerReportsDown(t *testing.T) {
	checker := NewDatabaseChecker(stubPinger{err: errors.New("dial tcp 127.0.0.1:5432: connect: connection refused")})

	if err := checker.Check(context.Background()); err == nil {
		t.Fatal("check returned nil, want an error for an unavailable database")
	}
}

func TestDatabaseCheckerHonoursTheCheckDeadline(t *testing.T) {
	checker := NewDatabaseChecker(slowPinger{})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := checker.Check(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("check returned %v, want %v", err, context.DeadlineExceeded)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("check blocked for %v, want a bounded check", elapsed)
	}
}

func TestNotificationCheckerReportsSubscriberAvailability(t *testing.T) {
	controller := gomock.NewController(t)
	reporter := mocks.NewMockHealthReporter(controller)
	gomock.InOrder(
		reporter.EXPECT().Healthy().Return(true),
		reporter.EXPECT().Healthy().Return(false),
	)

	checker := NewNotificationChecker("rabbitmq", reporter)
	if checker.Name() != "rabbitmq" {
		t.Fatalf("name = %q, want %q", checker.Name(), "rabbitmq")
	}
	if err := checker.Check(context.Background()); err != nil {
		t.Fatalf("check returned %v, want nil while the subscriber is connected", err)
	}
	if err := checker.Check(context.Background()); err == nil {
		t.Fatal("check returned nil, want an error while the subscriber is disconnected")
	}
}

func TestNotificationCheckerReportsDownWithoutAReporter(t *testing.T) {
	checker := NewNotificationChecker("rabbitmq", nil)

	if err := checker.Check(context.Background()); err == nil {
		t.Fatal("check returned nil, want an error when no broker health reporter is configured")
	}
}

func TestNotificationCheckerIgnoresPendingContextDeadline(t *testing.T) {
	checker := NewNotificationChecker("rabbitmq", stubHealthReporter{healthy: true})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := checker.Check(ctx); err != nil {
		t.Fatalf("check returned %v, want nil", err)
	}
}
