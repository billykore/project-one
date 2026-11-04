package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/billykore/project-one/internal/notifications/domain"
	pkgports "github.com/billykore/project-one/internal/pkg/ports"
	"github.com/billykore/project-one/internal/testkit/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type noopLogger struct{}

func (noopLogger) Debug(context.Context, string, ...any) {}
func (noopLogger) Info(context.Context, string, ...any)  {}
func (noopLogger) Warn(context.Context, string, ...any)  {}
func (noopLogger) Error(context.Context, string, ...any) {}
func (noopLogger) Fatal(context.Context, string, ...any) {}

// deliver runs the consumer against a subscriber that delivers payload once.
func deliver(t *testing.T, payload []byte, record func(*mocks.MockNotificationUseCase)) error {
	t.Helper()
	ctrl := gomock.NewController(t)
	subscriber := mocks.NewMockSubscriber(ctrl)
	uc := mocks.NewMockNotificationUseCase(ctrl)
	record(uc)
	subscriber.EXPECT().
		Subscribe(gomock.Any(), Topic, gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ string, handler pkgports.EventHandler) error {
			return handler(ctx, pkgports.Event{Payload: payload})
		})
	return NewConsumer(subscriber, uc, noopLogger{}).Run(context.Background())
}

func TestConsumer_PassesDecodedEventToUseCase(t *testing.T) {
	event := domain.NotificationEvent{
		EventID:       "backend-1",
		SchemaVersion: domain.NotificationEventSchemaVersion,
		Notification:  domain.Notification{UserID: 42, ActorID: 7, Type: domain.NotificationTypeFollow},
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	err = deliver(t, payload, func(uc *mocks.MockNotificationUseCase) {
		uc.EXPECT().RecordEvent(gomock.Any(), gomock.Cond(func(got domain.NotificationEvent) bool {
			return got.EventID == "backend-1" && got.Notification.UserID == 42
		})).Return(nil)
	})
	assert.NoError(t, err)
}

func TestConsumer_ReturnsUseCaseErrorForRetry(t *testing.T) {
	payload, err := json.Marshal(domain.NotificationEvent{EventID: "backend-2"})
	require.NoError(t, err)
	failure := errors.New("database unavailable")

	err = deliver(t, payload, func(uc *mocks.MockNotificationUseCase) {
		uc.EXPECT().RecordEvent(gomock.Any(), gomock.Any()).Return(failure)
	})
	assert.ErrorIs(t, err, failure)
}

func TestConsumer_RejectsMalformedPayload(t *testing.T) {
	err := deliver(t, []byte("{not json"), func(*mocks.MockNotificationUseCase) {})
	assert.Error(t, err)
}
