package pubsub

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/core/ports"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryPubSub(t *testing.T) {
	ctx := context.Background()

	t.Run("Publish and Subscribe", func(t *testing.T) {
		ps := NewInMemoryPubSub()
		defer func() { _ = ps.Close() }()

		var callCount int32
		var wg sync.WaitGroup
		wg.Add(2)

		handler1 := func(ctx context.Context, event ports.Event) error {
			atomic.AddInt32(&callCount, 1)
			assert.Equal(t, "test-payload", string(event.Payload))
			wg.Done()
			return nil
		}

		handler2 := func(ctx context.Context, event ports.Event) error {
			atomic.AddInt32(&callCount, 1)
			assert.Equal(t, "test-payload", string(event.Payload))
			wg.Done()
			return nil
		}

		publisher := NewInMemoryPublisher(ps)
		subscriber := NewInMemorySubscriber(ps)

		err := subscriber.Subscribe(ctx, "topic1", handler1)
		assert.NoError(t, err)

		err = subscriber.Subscribe(ctx, "topic1", handler2)
		assert.NoError(t, err)

		err = publisher.Publish(ctx, ports.Event{
			Topic:   "topic1",
			Payload: []byte("test-payload"),
		})
		assert.NoError(t, err)

		wg.Wait()
		assert.Equal(t, int32(2), atomic.LoadInt32(&callCount))
	})

	t.Run("Publish to Topic with No Subscribers", func(t *testing.T) {
		ps := NewInMemoryPubSub()
		defer func() { _ = ps.Close() }()

		publisher := NewInMemoryPublisher(ps)
		err := publisher.Publish(ctx, ports.Event{
			Topic:   "non-existent",
			Payload: []byte("no-one-cares"),
		})
		assert.NoError(t, err)
	})

	t.Run("Operations After Close", func(t *testing.T) {
		ps := NewInMemoryPubSub()
		publisher := NewInMemoryPublisher(ps)
		subscriber := NewInMemorySubscriber(ps)

		err := ps.Close()
		assert.NoError(t, err)

		// Publish should return error after close
		err = publisher.Publish(ctx, ports.Event{Topic: "topic1"})
		assert.True(t, errors.Is(err, ErrPubSubClosed))

		// Subscribe should return error after close
		err = subscriber.Subscribe(ctx, "topic1", func(ctx context.Context, event ports.Event) error {
			return nil
		})
		assert.True(t, errors.Is(err, ErrPubSubClosed))
	})

	t.Run("Close Waits for In-flight Handlers", func(t *testing.T) {
		ps := NewInMemoryPubSub()
		publisher := NewInMemoryPublisher(ps)
		subscriber := NewInMemorySubscriber(ps)

		started := make(chan struct{})
		done := make(chan struct{})

		handler := func(ctx context.Context, event ports.Event) error {
			close(started)
			time.Sleep(100 * time.Millisecond)
			close(done)
			return nil
		}

		err := subscriber.Subscribe(ctx, "topic1", handler)
		assert.NoError(t, err)

		err = publisher.Publish(ctx, ports.Event{Topic: "topic1"})
		assert.NoError(t, err)

		// Wait until handler starts
		<-started

		// Close should wait for handler to finish
		err = ps.Close()
		assert.NoError(t, err)

		// Handler should be done by the time Close returns
		select {
		case <-done:
			// Success
		default:
			t.Fatal("expected Close to block until handler completed")
		}
	})
}
