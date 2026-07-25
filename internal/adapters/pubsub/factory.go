package pubsub

import (
	"fmt"
	"sync"

	"github.com/billykore/project-one/internal/config"
	"github.com/billykore/project-one/internal/core/ports"
)

var (
	inMemoryOnce sync.Once
	inMemoryPS   *inMemoryPubSub
)

func sharedInMemoryPubSub() *inMemoryPubSub {
	inMemoryOnce.Do(func() {
		inMemoryPS = NewInMemoryPubSub()
	})
	return inMemoryPS
}

// NewPublisher creates a ports.Publisher based on the provided broker configuration.
// Supported types: "kafka", "rabbitmq", "inmemory" (default).
func NewPublisher(cfg config.MessageBrokerConfig, logger ports.Logger) (ports.Publisher, error) {
	switch cfg.Type {
	case "kafka":
		return NewKafkaPublisher(cfg.Kafka, logger)
	case "rabbitmq":
		return NewRabbitMQPublisher(cfg.RabbitMQ, logger)
	case "inmemory", "":
		return NewInMemoryPublisher(sharedInMemoryPubSub()), nil
	default:
		return nil, fmt.Errorf("unknown message broker type: %q", cfg.Type)
	}
}

// NewSubscriber creates a ports.Subscriber based on the provided broker configuration.
// Supported types: "kafka", "rabbitmq", "inmemory" (default).
func NewSubscriber(cfg config.MessageBrokerConfig, logger ports.Logger) (ports.Subscriber, error) {
	switch cfg.Type {
	case "kafka":
		return NewKafkaSubscriber(cfg.Kafka, logger)
	case "rabbitmq":
		return NewRabbitMQSubscriber(cfg.RabbitMQ, logger)
	case "inmemory", "":
		return NewInMemorySubscriber(sharedInMemoryPubSub()), nil
	default:
		return nil, fmt.Errorf("unknown message broker type: %q", cfg.Type)
	}
}
