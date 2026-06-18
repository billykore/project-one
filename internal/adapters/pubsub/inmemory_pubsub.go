package pubsub

import (
	"context"
	"sync"

	"github.com/billykore/project-one/internal/core/ports"
)

// inMemoryPubSub is a thread-safe, in-memory publish/subscribe broker.
// It implements both ports.Publisher and ports.Subscriber.
type inMemoryPubSub struct {
	mu     sync.RWMutex
	subs   map[string][]ports.EventHandler
	closed bool
	wg     sync.WaitGroup
}

// NewInMemoryPublisher creates a new in-memory Publisher.
func NewInMemoryPublisher(ps *inMemoryPubSub) ports.Publisher {
	return ps
}

// NewInMemorySubscriber creates a new in-memory Subscriber.
func NewInMemorySubscriber(ps *inMemoryPubSub) ports.Subscriber {
	return ps
}

// NewInMemoryPubSub creates the shared in-memory broker instance.
// Pass the returned value to NewInMemoryPublisher and NewInMemorySubscriber.
func NewInMemoryPubSub() *inMemoryPubSub {
	return &inMemoryPubSub{
		subs: make(map[string][]ports.EventHandler),
	}
}

func (ps *inMemoryPubSub) Publish(ctx context.Context, event ports.Event) error {
	ps.mu.RLock()
	if ps.closed {
		ps.mu.RUnlock()
		return ErrPubSubClosed
	}
	handlers := make([]ports.EventHandler, len(ps.subs[event.Topic]))
	copy(handlers, ps.subs[event.Topic])
	ps.mu.RUnlock()

	for _, h := range handlers {
		ps.wg.Add(1)
		go func(handler ports.EventHandler) {
			defer ps.wg.Done()
			// Errors from individual handlers are intentionally ignored
			// to avoid blocking other subscribers on the same topic.
			_ = handler(ctx, event)
		}(h)
	}

	return nil
}

func (ps *inMemoryPubSub) Subscribe(_ context.Context, topic string, handler ports.EventHandler) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return ErrPubSubClosed
	}

	ps.subs[topic] = append(ps.subs[topic], handler)
	return nil
}

func (ps *inMemoryPubSub) Close() error {
	ps.mu.Lock()
	if ps.closed {
		ps.mu.Unlock()
		return nil
	}
	ps.closed = true
	ps.mu.Unlock()

	// Wait for all in-flight handlers to complete.
	ps.wg.Wait()
	return nil
}
