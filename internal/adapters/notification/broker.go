package notification

import (
	"context"
	"sync"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

var _ ports.NotificationPublisher = (*MemoryBroker)(nil)

type MemoryBroker struct {
	ch   chan *domain.Notification
	mu   sync.Mutex
	done bool
}

func NewMemoryBroker(bufferSize int) *MemoryBroker {
	return &MemoryBroker{
		ch: make(chan *domain.Notification, bufferSize),
	}
}

func (b *MemoryBroker) Publish(ctx context.Context, notification *domain.Notification) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.done {
		return context.Canceled
	}
	select {
	case b.ch <- notification:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *MemoryBroker) Channel() <-chan *domain.Notification {
	return b.ch
}

func (b *MemoryBroker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.done {
		b.done = true
		close(b.ch)
	}
}
