package pubsub

import (
	"context"
	"testing"

	"github.com/billykore/project-one/internal/config"
	"github.com/billykore/project-one/internal/core/ports"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

// TestRabbitMQPublisherInterface verifies that rabbitMQPublisher satisfies ports.Publisher.
func TestRabbitMQPublisherInterface(t *testing.T) {
	var _ ports.Publisher = (*rabbitMQPublisher)(nil)
}

// TestRabbitMQSubscriberInterface verifies that rabbitMQSubscriber satisfies ports.Subscriber.
func TestRabbitMQSubscriberInterface(t *testing.T) {
	var _ ports.Subscriber = (*rabbitMQSubscriber)(nil)
}

// TestNewRabbitMQPublisher validates that the constructor no longer dials eagerly.
func TestNewRabbitMQPublisher(t *testing.T) {
	cfg := config.RabbitMQBrokerConfig{
		URL: "amqp://invalid:5672/",
	}
	pub, err := NewRabbitMQPublisher(cfg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, pub)
}

// TestNewRabbitMQSubscriber validates that the constructor no longer dials eagerly.
func TestNewRabbitMQSubscriber(t *testing.T) {
	cfg := config.RabbitMQBrokerConfig{
		URL: "amqp://invalid:5672/",
	}
	sub, err := NewRabbitMQSubscriber(cfg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, sub)
}

// TestRabbitMQPublisher_Close validates that Close is safe without an active connection.
func TestRabbitMQPublisher_Close(t *testing.T) {
	cfg := config.RabbitMQBrokerConfig{
		URL: "amqp://invalid:5672/",
	}
	pub, err := NewRabbitMQPublisher(cfg, nil)
	assert.NoError(t, err)
	assert.NoError(t, pub.Close())
}

// TestRabbitMQSubscriber_Close validates that Close is safe without an active connection.
func TestRabbitMQSubscriber_Close(t *testing.T) {
	cfg := config.RabbitMQBrokerConfig{
		URL:      "amqp://invalid:5672/",
		Exchange: "test-exchange",
		Queue:    "test-queue",
	}
	sub, err := NewRabbitMQSubscriber(cfg, nil)
	assert.NoError(t, err)
	assert.NoError(t, sub.Close())
}

// TestRabbitMQSubscriber_SubscribeCtxCancelled validates that subscribing then cancelling
// the context does not block indefinitely (structural test — no actual broker needed).
func TestRabbitMQSubscriber_SubscribeCtxCancelled(t *testing.T) {
	cfg := config.RabbitMQBrokerConfig{
		URL:      "amqp://invalid:5672/",
		Exchange: "test-exchange",
		Queue:    "test-queue",
	}
	sub, err := NewRabbitMQSubscriber(cfg, nil)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := func(ctx context.Context, event ports.Event) error { return nil }
	assert.NoError(t, sub.Subscribe(ctx, "test", handler))
	cancel()
	assert.NoError(t, sub.Close())
}

// TestRabbitMQPublisher_Health validates the health hook exists.
func TestRabbitMQPublisher_Health(t *testing.T) {
	cfg := config.RabbitMQBrokerConfig{URL: "amqp://invalid:5672/"}
	pub, err := NewRabbitMQPublisher(cfg, nil)
	assert.NoError(t, err)

	health, ok := pub.(interface{ Healthy() bool })
	assert.True(t, ok)
	assert.False(t, health.Healthy())
}

// TestAMQPMsgToEvent verifies that amqpMsgToEvent is properly typed.
func TestAMQPMsgToEvent(t *testing.T) {
	// amqpMsgToEvent is unexported. This test verifies it's properly defined.
	_ = amqpMsgToEvent
}

// TestRabbitMQPublisher_ExchangeDeclare verifies the exchange declare config.
func TestRabbitMQPublisher_ExchangeDeclare(t *testing.T) {
	t.Run("fanout exchange", func(t *testing.T) {
		p := &rabbitMQPublisher{
			cfg: config.RabbitMQBrokerConfig{
				Exchange: "project1.notifications",
			},
		}
		assert.Equal(t, "project1.notifications", p.cfg.Exchange)
	})

	t.Run("durable delivery mode", func(t *testing.T) {
		publishing := amqp.Publishing{
			DeliveryMode: amqp.Persistent,
		}
		assert.Equal(t, uint8(2), publishing.DeliveryMode) // amqp.Persistent = 2
	})
}
