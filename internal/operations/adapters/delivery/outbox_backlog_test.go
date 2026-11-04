package delivery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/operations/domain"
	"github.com/billykore/project-one/internal/pkg/outbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOutboxStats struct {
	stats outbox.Stats
	err   error
}

func (f fakeOutboxStats) Stats(context.Context) (outbox.Stats, error) { return f.stats, f.err }

func TestOutboxBacklog_ProjectsStats(t *testing.T) {
	backlog, err := NewOutboxBacklog(fakeOutboxStats{stats: outbox.Stats{Pending: 4, OldestPendingAge: 90 * time.Second, Failed: 2}}).
		DeliveryBacklog(context.Background())

	require.NoError(t, err)
	assert.Equal(t, domain.DeliveryBacklog{Pending: 4, OldestPendingAge: 90 * time.Second, Failed: 2}, backlog)
}

func TestOutboxBacklog_PassesThroughErrors(t *testing.T) {
	boom := errors.New("boom")

	_, err := NewOutboxBacklog(fakeOutboxStats{err: boom}).DeliveryBacklog(context.Background())

	assert.ErrorIs(t, err, boom)
}
