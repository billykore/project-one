// Package delivery adapts the transactional outbox to operations' backlog port.
package delivery

import (
	"context"

	"github.com/billykore/project-one/internal/operations/domain"
	"github.com/billykore/project-one/internal/operations/ports"
	"github.com/billykore/project-one/internal/pkg/outbox"
)

// outboxStats is the slice of the outbox stats reader this adapter uses.
type outboxStats interface {
	Stats(ctx context.Context) (outbox.Stats, error)
}

type outboxBacklog struct {
	stats outboxStats
}

var _ ports.DeliveryBacklogReader = (*outboxBacklog)(nil)

// NewOutboxBacklog adapts outbox statistics to the DeliveryBacklogReader port.
func NewOutboxBacklog(stats outboxStats) ports.DeliveryBacklogReader {
	return &outboxBacklog{stats: stats}
}

// DeliveryBacklog returns the operations projection of the outbox statistics.
func (b *outboxBacklog) DeliveryBacklog(ctx context.Context) (domain.DeliveryBacklog, error) {
	stats, err := b.stats.Stats(ctx)
	if err != nil {
		return domain.DeliveryBacklog{}, err
	}
	return domain.DeliveryBacklog{
		Pending:          stats.Pending,
		OldestPendingAge: stats.OldestPendingAge,
		Failed:           stats.Failed,
	}, nil
}
