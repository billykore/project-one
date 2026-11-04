// Package consumer is the broker-driven entry point of the notifications context.
package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/billykore/project-one/internal/notifications/domain"
	"github.com/billykore/project-one/internal/notifications/ports"
	pkgports "github.com/billykore/project-one/internal/pkg/ports"
)

// Topic is the broker topic that carries notification events.
const Topic = "notifications"

// Consumer decodes notification events from the broker and hands them to the
// notification use case.
type Consumer struct {
	subscriber pkgports.Subscriber
	uc         ports.NotificationUseCase
	log        pkgports.Logger
}

// NewConsumer creates a notification event consumer.
func NewConsumer(subscriber pkgports.Subscriber, uc ports.NotificationUseCase, log pkgports.Logger) *Consumer {
	return &Consumer{subscriber: subscriber, uc: uc, log: log}
}

// Run subscribes to the notification topic and blocks until the subscription ends.
func (c *Consumer) Run(ctx context.Context) error {
	return c.subscriber.Subscribe(ctx, Topic, c.handle)
}

func (c *Consumer) handle(ctx context.Context, event pkgports.Event) error {
	var notificationEvent domain.NotificationEvent
	if err := json.Unmarshal(event.Payload, &notificationEvent); err != nil {
		c.log.Error(ctx, "failed to unmarshal notification event", "error", err)
		return fmt.Errorf("decode notification event: %w", err)
	}
	if err := c.uc.RecordEvent(ctx, notificationEvent); err != nil {
		c.log.Error(ctx, "failed to record notification event", "event_id", notificationEvent.EventID, "error", err)
		return err
	}
	return nil
}
