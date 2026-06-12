package notification

import (
	"context"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

type mockRepo struct {
	notifications []*domain.Notification
}

func (r *mockRepo) Create(ctx context.Context, n *domain.Notification) error {
	r.notifications = append(r.notifications, n)
	return nil
}
func (r *mockRepo) GetByID(ctx context.Context, id int) (*domain.Notification, error) {
	return nil, nil
}
func (r *mockRepo) GetByUserID(ctx context.Context, u int, l, o int) ([]*domain.Notification, error) {
	return nil, nil
}
func (r *mockRepo) MarkAsRead(ctx context.Context, id int) error   { return nil }
func (r *mockRepo) MarkAllAsRead(ctx context.Context, u int) error { return nil }

type mockLogger struct{}

func (mockLogger) Debug(ctx context.Context, msg string, fields ...any) {}
func (mockLogger) Info(ctx context.Context, msg string, fields ...any)  {}
func (mockLogger) Warn(ctx context.Context, msg string, fields ...any)  {}
func (mockLogger) Error(ctx context.Context, msg string, fields ...any) {}
func (mockLogger) Fatal(ctx context.Context, msg string, fields ...any) {}

func TestBrokerAndWorker(t *testing.T) {
	broker := NewMemoryBroker(10)
	repo := &mockRepo{}
	worker := NewBackgroundWorker(repo, broker.Channel(), mockLogger{})

	ctx := context.Background()
	err := worker.Start(ctx)
	assert.NoError(t, err)

	n := &domain.Notification{UserID: 1, ActorID: 2, Type: domain.NotificationTypeFollow}
	err = broker.Publish(ctx, n)
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	assert.Len(t, repo.notifications, 1)

	err = worker.Stop(ctx)
	assert.NoError(t, err)
	broker.Close()
}
