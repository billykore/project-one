package outbox

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	platformports "github.com/billykore/project-one/internal/platform/ports"
	"github.com/billykore/project-one/internal/testkit/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPublisherAndDispatcherPersistAndDeliverEvents(t *testing.T) {
	dsn := os.Getenv("OUTBOX_TEST_DSN")
	if dsn == "" {
		t.Skip("OUTBOX_TEST_DSN is not configured")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer func() { require.NoError(t, sqlDB.Close()) }()

	ctx := context.Background()
	eventID := "outbox-integration-" + time.Now().UTC().Format("20060102150405.000000000")
	t.Cleanup(func() { _ = db.Where("event_id = ?", eventID).Delete(&messageModel{}).Error })

	persistentPublisher := NewPublisher(db)
	event := platformports.Event{Topic: "notifications", Key: "user:42", Payload: []byte(`{"event":"test"}`), Metadata: map[string]string{"event_id": eventID, "schema_version": "1.0"}}
	require.NoError(t, persistentPublisher.InTransaction(ctx, func(txCtx context.Context) error {
		return persistentPublisher.Publish(txCtx, event)
	}))

	var stored messageModel
	require.NoError(t, db.Where("event_id = ?", eventID).Take(&stored).Error)
	require.Nil(t, stored.PublishedAt)

	ctrl := gomock.NewController(t)
	broker := mocks.NewMockPublisher(ctrl)
	log := mocks.NewMockLogger(ctrl)
	broker.EXPECT().Publish(gomock.Any(), event).Return(nil)
	dispatcher := NewDispatcher(db, broker, log, time.Hour, 1)
	require.NoError(t, dispatcher.DispatchOnce(ctx))

	require.NoError(t, db.Where("event_id = ?", eventID).Take(&stored).Error)
	require.NotNil(t, stored.PublishedAt)

	rollbackID := eventID + "-rollback"
	rollbackEvent := event
	rollbackEvent.Metadata = map[string]string{"event_id": rollbackID}
	err = persistentPublisher.InTransaction(ctx, func(txCtx context.Context) error {
		require.NoError(t, persistentPublisher.Publish(txCtx, rollbackEvent))
		return errors.New("force rollback")
	})
	require.Error(t, err)
	var count int64
	require.NoError(t, db.Model(&messageModel{}).Where("event_id = ?", rollbackID).Count(&count).Error)
	require.Zero(t, count)
}
