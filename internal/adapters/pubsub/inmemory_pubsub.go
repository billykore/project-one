package pubsub

import (
	"context"
	"slices"
	"sync"

	"github.com/billykore/project-one/internal/core/ports"
)

// ponytail: simplified inMemoryPubSub broker implementing ports.Publisher and ports.Subscriber.
type inMemoryPubSub struct {
	mu     sync.RWMutex
	subs   map[string][]ports.EventHandler
	closed bool
	wg     sync.WaitGroup
}

// NewInMemoryPublisher creates a new in-memory Publisher.
func NewInMemoryPublisher(ps *inMemoryPubSub) ports.Publisher { return ps }

// NewInMemorySubscriber creates a new in-memory Subscriber.
func NewInMemorySubscriber(ps *inMemoryPubSub) ports.Subscriber { return ps }

// NewInMemoryPubSub creates the shared in-memory broker instance.
func NewInMemoryPubSub() *inMemoryPubSub {
	return &inMemoryPubSub{subs: make(map[string][]ports.EventHandler)}
}

// Publish publishes an event concurrently to all subscribers registered on the topic.
func (ps *inMemoryPubSub) Publish(ctx context.Context, event ports.Event) error {
	ps.mu.RLock()
	closed := ps.closed
	var handlers []ports.EventHandler
	if !closed {
		// ponytail: use slices.Clone from stdlib for safe, allocation-optimized copy
		handlers = slices.Clone(ps.subs[event.Topic])
	}
	ps.mu.RUnlock()

	if closed {
		return ErrPubSubClosed
	}

	for _, h := range handlers {
		ps.wg.Add(1)
		go func(handler ports.EventHandler) {
			defer ps.wg.Done()
			_ = handler(ctx, event)
		}(h)
	}
	return nil
}

// Subscribe registers an event handler on the specified topic.
func (ps *inMemoryPubSub) Subscribe(_ context.Context, topic string, handler ports.EventHandler) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if ps.closed {
		return ErrPubSubClosed
	}
	ps.subs[topic] = append(ps.subs[topic], handler)
	return nil
}

// Close gracefully closes the pubsub broker, waiting for in-flight handlers to finish.
func (ps *inMemoryPubSub) Close() error {
	ps.mu.Lock()
	if ps.closed {
		ps.mu.Unlock()
		return nil
	}
	ps.closed = true
	ps.mu.Unlock()

	ps.wg.Wait()
	return nil
}
