package outbox

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/testkit/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// openTestDB connects to the database named by OUTBOX_TEST_DSN, which must
// have every migration applied, and skips the test when it is not configured.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("OUTBOX_TEST_DSN")
	if dsn == "" {
		t.Skip("OUTBOX_TEST_DSN is not configured")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

// insertMessage stores a message row directly so a test controls its state.
func insertMessage(t *testing.T, db *gorm.DB, model messageModel) messageModel {
	t.Helper()
	if model.Topic == "" {
		model.Topic = "notifications"
	}
	if model.MessageKey == "" {
		model.MessageKey = "user:42"
	}
	if model.Payload == nil {
		model.Payload = []byte(`{}`)
	}
	if model.Metadata == "" {
		model.Metadata = `{"event_id":"` + model.EventID + `"}`
	}
	if model.AvailableAt.IsZero() {
		model.AvailableAt = time.Now().UTC().Add(-time.Second)
	}
	if model.CreatedAt.IsZero() {
		model.CreatedAt = time.Now().UTC()
	}
	require.NoError(t, db.Create(&model).Error)
	t.Cleanup(func() { _ = db.Where("id = ?", model.ID).Delete(&messageModel{}).Error })
	return model
}

func uniqueEventID(prefix string) string {
	return prefix + "-" + time.Now().UTC().Format("20060102150405.000000000")
}

func TestDispatcherParksMessageAfterMaxAttempts(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	stored := insertMessage(t, db, messageModel{EventID: uniqueEventID("outbox-failed"), Attempts: 2})

	ctrl := gomock.NewController(t)
	broker := mocks.NewMockPublisher(ctrl)
	log := mocks.NewMockLogger(ctrl)
	broker.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(errors.New("broker unavailable"))
	log.EXPECT().Warn(gomock.Any(), "outbox publish failed; event will be retried", gomock.Any()).Times(1)
	log.EXPECT().Error(gomock.Any(), "outbox message exhausted delivery attempts; parked as failed", "event_id", stored.EventID, "attempts", 3)

	dispatcher := NewDispatcher(db, broker, log, time.Hour, 10, WithMaxAttempts(3))
	require.NoError(t, dispatcher.DispatchOnce(ctx))

	var got messageModel
	require.NoError(t, db.Where("id = ?", stored.ID).Take(&got).Error)
	assert.Equal(t, 3, got.Attempts)
	assert.NotNil(t, got.FailedAt)
	assert.Nil(t, got.PublishedAt)
	assert.Nil(t, got.LockedUntil)

	// A parked message is never claimed again, so the broker is not called.
	require.NoError(t, dispatcher.DispatchOnce(ctx))
}

func TestDispatcherRetriesBeforeMaxAttempts(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	stored := insertMessage(t, db, messageModel{EventID: uniqueEventID("outbox-retry"), Attempts: 1})

	ctrl := gomock.NewController(t)
	broker := mocks.NewMockPublisher(ctrl)
	log := mocks.NewMockLogger(ctrl)
	broker.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(errors.New("broker unavailable"))
	log.EXPECT().Warn(gomock.Any(), "outbox publish failed; event will be retried", gomock.Any())

	before := time.Now().UTC()
	require.NoError(t, NewDispatcher(db, broker, log, time.Hour, 10, WithMaxAttempts(3)).DispatchOnce(ctx))

	var got messageModel
	require.NoError(t, db.Where("id = ?", stored.ID).Take(&got).Error)
	assert.Equal(t, 2, got.Attempts)
	assert.Nil(t, got.FailedAt)
	// Attempt 1 backs off for 2 seconds.
	assert.WithinDuration(t, before.Add(2*time.Second), got.AvailableAt, time.Second)
}

func TestDispatcherCleanupDeletesOnlyExpiredPublishedMessages(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	old := now.Add(-48 * time.Hour)
	recent := now.Add(-time.Hour)

	expired := insertMessage(t, db, messageModel{EventID: uniqueEventID("outbox-expired"), PublishedAt: &old})
	kept := insertMessage(t, db, messageModel{EventID: uniqueEventID("outbox-recent"), PublishedAt: &recent})
	parked := insertMessage(t, db, messageModel{EventID: uniqueEventID("outbox-parked"), FailedAt: &old, CreatedAt: old})
	pending := insertMessage(t, db, messageModel{EventID: uniqueEventID("outbox-pending"), CreatedAt: old})

	ctrl := gomock.NewController(t)
	dispatcher := NewDispatcher(db, mocks.NewMockPublisher(ctrl), mocks.NewMockLogger(ctrl), time.Hour, 10, WithRetention(24*time.Hour))
	deleted, err := dispatcher.Cleanup(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, deleted, int64(1))

	exists := func(id int64) bool {
		var count int64
		require.NoError(t, db.Model(&messageModel{}).Where("id = ?", id).Count(&count).Error)
		return count == 1
	}
	assert.False(t, exists(expired.ID), "published message past retention is deleted")
	assert.True(t, exists(kept.ID), "published message within retention is kept")
	assert.True(t, exists(parked.ID), "failed message is kept for an operator")
	assert.True(t, exists(pending.ID), "undelivered message is kept")
}

func TestStatsReaderReportsBacklog(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	reader := NewStatsReader(db)

	baseline, err := reader.Stats(ctx)
	require.NoError(t, err)

	now := time.Now().UTC()
	published := now.Add(-time.Minute)
	failed := now.Add(-time.Minute)
	insertMessage(t, db, messageModel{EventID: uniqueEventID("stats-pending-old"), CreatedAt: now.Add(-90 * time.Minute)})
	insertMessage(t, db, messageModel{EventID: uniqueEventID("stats-pending-new"), CreatedAt: now})
	insertMessage(t, db, messageModel{EventID: uniqueEventID("stats-failed"), FailedAt: &failed, CreatedAt: now.Add(-10 * time.Hour)})
	insertMessage(t, db, messageModel{EventID: uniqueEventID("stats-published"), PublishedAt: &published, CreatedAt: now.Add(-20 * time.Hour)})

	stats, err := reader.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, baseline.Pending+2, stats.Pending)
	assert.Equal(t, baseline.Failed+1, stats.Failed)
	// The oldest pending message is 90 minutes old; failed and published
	// messages do not count toward the age.
	assert.GreaterOrEqual(t, stats.OldestPendingAge, 89*time.Minute)
	if baseline.Pending == 0 {
		assert.Less(t, stats.OldestPendingAge, 2*time.Hour)
	}
}
