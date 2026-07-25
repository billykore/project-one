package pubsub

import (
	"context"
	"slices"
	"sync"

	"github.com/billykore/project-one/internal/core/ports"
)

type inMemoryPubSub struct {
	mu      sync.RWMutex
	subs    map[string][]ports.EventHandler
	closed  bool
	wg      sync.WaitGroup
	workers map[string]*inMemoryWorker
}

type inMemoryWorker struct {
	once   sync.Once
	queue  chan inMemoryJob
	closed chan struct{}
}

type inMemoryJob struct {
	ctx   context.Context
	event ports.Event
}

// NewInMemoryPublisher creates a new in-memory Publisher.
func NewInMemoryPublisher(ps *inMemoryPubSub) ports.Publisher { return ps }

// NewInMemorySubscriber creates a new in-memory Subscriber.
func NewInMemorySubscriber(ps *inMemoryPubSub) ports.Subscriber { return ps }

// NewInMemoryPubSub creates the shared in-memory broker instance.
func NewInMemoryPubSub() *inMemoryPubSub {
	return &inMemoryPubSub{
		subs:    make(map[string][]ports.EventHandler),
		workers: make(map[string]*inMemoryWorker),
	}
}

// Publish publishes an event concurrently to all subscribers registered on the topic.
func (ps *inMemoryPubSub) Publish(ctx context.Context, event ports.Event) error {
	ps.mu.RLock()
	closed := ps.closed
	handlers := slices.Clone(ps.subs[event.Topic])
	if closed {
		ps.mu.RUnlock()
		return ErrPubSubClosed
	}
	if len(handlers) == 0 {
		ps.mu.RUnlock()
		return nil
	}
	worker := ps.workerForLocked(event.Topic, event.Key)

	ps.wg.Add(1)
	select {
	case worker.queue <- inMemoryJob{ctx: ctx, event: event}:
		ps.mu.RUnlock()
		return nil
	case <-ctx.Done():
		ps.wg.Done()
		ps.mu.RUnlock()
		return ctx.Err()
	}
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
	workers := make([]*inMemoryWorker, 0, len(ps.workers))
	for _, worker := range ps.workers {
		workers = append(workers, worker)
	}
	ps.mu.Unlock()

	for _, worker := range workers {
		worker.once.Do(func() {
			close(worker.queue)
			close(worker.closed)
		})
	}
	ps.wg.Wait()
	return nil
}

func (ps *inMemoryPubSub) Healthy() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return !ps.closed
}

func (ps *inMemoryPubSub) workerForLocked(topic, key string) *inMemoryWorker {
	if ps.workers == nil {
		ps.workers = make(map[string]*inMemoryWorker)
	}

	dispatchKey := topic + "\x00" + key
	if worker, ok := ps.workers[dispatchKey]; ok {
		return worker
	}

	worker := &inMemoryWorker{
		queue:  make(chan inMemoryJob, 64),
		closed: make(chan struct{}),
	}
	ps.workers[dispatchKey] = worker

	ps.wg.Add(1)
	go func(topic string, worker *inMemoryWorker) {
		defer ps.wg.Done()
		for job := range worker.queue {
			ps.mu.RLock()
			handlers := slices.Clone(ps.subs[topic])
			ps.mu.RUnlock()

			for _, handler := range handlers {
				_ = handler(job.ctx, job.event)
			}
			ps.wg.Done()
		}
	}(topic, worker)

	return worker
}
