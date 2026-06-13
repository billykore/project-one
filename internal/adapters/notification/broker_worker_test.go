package notification

import (
	"context"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

type mockLogger struct{}

func (mockLogger) Debug(ctx context.Context, msg string, fields ...any) {}
func (mockLogger) Info(ctx context.Context, msg string, fields ...any)  {}
func (mockLogger) Warn(ctx context.Context, msg string, fields ...any)  {}
func (mockLogger) Error(ctx context.Context, msg string, fields ...any) {}
func (mockLogger) Fatal(ctx context.Context, msg string, fields ...any) {}

func TestBrokerAndWorker(t *testing.T) {
	broker := NewMemoryBroker(10)
	worker := NewBackgroundWorker(broker.Channel(), mockLogger{})

	ctx := context.Background()
	outCh, err := worker.Start(ctx)
	assert.NoError(t, err)

	n := &domain.Notification{UserID: 1, ActorID: 2, Type: domain.NotificationTypeFollow}
	err = broker.Publish(ctx, n)
	assert.NoError(t, err)

	// Consume and save notifications in a mock consumer routine
	go func() {
		for notification := range outCh {
			t.Logf("Received notification for user %d of type %s", notification.UserID, notification.Type)
			t.Log(notification)
		}
	}()

	time.Sleep(1 * time.Second)

	broker.Close()
	err = worker.Stop(ctx)
	assert.NoError(t, err)
}
