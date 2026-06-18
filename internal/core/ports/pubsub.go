package ports

import "context"

// Event represents a message published to a topic.
type Event struct {
	// Topic is the subject or channel the event belongs to.
	Topic string
	// Key is an optional partitioning key for ordered delivery.
	Key string
	// Payload is the serialized event data.
	Payload []byte
	// Metadata holds optional key-value pairs for headers or tracing information.
	Metadata map[string]string
}

// Publisher is a driven port for publishing events to a message broker.
type Publisher interface {
	// Publish sends an event to the specified topic.
	Publish(ctx context.Context, event Event) error
	// Close gracefully shuts down the publisher, flushing any pending messages.
	Close() error
}

// Subscriber is a driven port for consuming events from a message broker.
type Subscriber interface {
	// Subscribe registers a handler for the given topic.
	// The handler is invoked for each incoming event.
	// It returns an error if the subscription cannot be established.
	Subscribe(ctx context.Context, topic string, handler EventHandler) error
	// Close gracefully shuts down the subscriber, releasing all resources.
	Close() error
}

// EventHandler is a callback function invoked when an event is received.
// Returning a non-nil error signals that the event was not processed successfully
// and should be retried or dead-lettered depending on the implementation.
type EventHandler func(ctx context.Context, event Event) error
