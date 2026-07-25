package pubsub

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/billykore/project-one/internal/config"
	"github.com/billykore/project-one/internal/core/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

const rabbitMQDeadLetterSuffix = ".dlq"
const rabbitMQRetryAttempts = 3

// rabbitMQPublisher implements ports.Publisher using rabbitmq/amqp091-go.
type rabbitMQPublisher struct {
	cfg      config.RabbitMQBrokerConfig
	log      ports.Logger
	conn     *amqp.Connection
	channel  *amqp.Channel
	confirms <-chan amqp.Confirmation
	mu       sync.Mutex
	healthy  atomic.Bool
}

// NewRabbitMQPublisher creates a RabbitMQ-backed ports.Publisher.
func NewRabbitMQPublisher(cfg config.RabbitMQBrokerConfig, logger ports.Logger) (ports.Publisher, error) {
	p := &rabbitMQPublisher{cfg: cfg, log: logger}
	p.healthy.Store(false)
	return p, nil
}

func (p *rabbitMQPublisher) connect(ctx context.Context) error {
	if p.conn != nil && p.channel != nil {
		return nil
	}

	conn, ch, err := dialRabbitMQ(ctx, p.cfg, p.log)
	if err != nil {
		return err
	}
	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("rabbitmq confirm: %w", err)
	}

	p.conn = conn
	p.channel = ch
	p.confirms = ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	p.healthy.Store(true)
	return nil
}

// Publish publishes a notification event to the configured exchange.
func (p *rabbitMQPublisher) Publish(ctx context.Context, event ports.Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.connect(ctx); err != nil {
		p.healthy.Store(false)
		p.log.Warn(ctx, "rabbitmq publish connect failed", "exchange", p.cfg.Exchange, "error", err)
		return fmt.Errorf("rabbitmq connect: %w", err)
	}

	headers := amqp.Table{}
	for k, v := range event.Metadata {
		headers[k] = v
	}

	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := p.channel.PublishWithContext(publishCtx,
		p.cfg.Exchange,
		event.Key,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         event.Payload,
			DeliveryMode: amqp.Persistent,
			Headers:      headers,
			MessageId:    event.Metadata["event_id"],
		},
	)
	if err != nil {
		p.healthy.Store(false)
		if p.log != nil {
			p.log.Warn(ctx, "rabbitmq publish failed", "exchange", p.cfg.Exchange, "error", err)
		}
		_ = p.closeLocked()
		return fmt.Errorf("rabbitmq publish: %w", err)
	}
	select {
	case confirmation := <-p.confirms:
		if !confirmation.Ack {
			p.healthy.Store(false)
			_ = p.closeLocked()
			return fmt.Errorf("rabbitmq publish nacked by broker")
		}
	case <-publishCtx.Done():
		p.healthy.Store(false)
		_ = p.closeLocked()
		return fmt.Errorf("rabbitmq publish confirm timeout: %w", publishCtx.Err())
	}

	p.healthy.Store(true)
	if p.log != nil {
		p.log.Debug(ctx, "rabbitmq message published", "exchange", p.cfg.Exchange, "key", event.Key)
	}
	return nil
}

// Close gracefully closes the RabbitMQ connection.
func (p *rabbitMQPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.log != nil {
		p.log.Info(context.Background(), "closing rabbitmq publisher")
	}
	p.healthy.Store(false)
	return p.closeLocked()
}

func (p *rabbitMQPublisher) closeLocked() error {
	var err error
	if p.channel != nil {
		if closeErr := p.channel.Close(); closeErr != nil {
			err = closeErr
		}
		p.channel = nil
		p.confirms = nil
	}
	if p.conn != nil {
		if closeErr := p.conn.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		p.conn = nil
	}
	return err
}

func (p *rabbitMQPublisher) Healthy() bool {
	return p.healthy.Load()
}

// rabbitMQSubscriber implements ports.Subscriber using rabbitmq/amqp091-go.
type rabbitMQSubscriber struct {
	cfg     config.RabbitMQBrokerConfig
	log     ports.Logger
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.Mutex
	wg      sync.WaitGroup
	healthy atomic.Bool
}

// NewRabbitMQSubscriber creates a RabbitMQ-backed ports.Subscriber.
func NewRabbitMQSubscriber(cfg config.RabbitMQBrokerConfig, logger ports.Logger) (ports.Subscriber, error) {
	if logger != nil {
		logger.Info(context.Background(), "rabbitmq subscriber created", "exchange", cfg.Exchange, "queue", cfg.Queue)
	}
	s := &rabbitMQSubscriber{cfg: cfg, log: logger}
	s.healthy.Store(false)
	return s, nil
}

func (s *rabbitMQSubscriber) connect(ctx context.Context) error {
	if s.conn != nil && s.channel != nil {
		return nil
	}

	conn, ch, err := dialRabbitMQ(ctx, s.cfg, s.log)
	if err != nil {
		return err
	}

	dlqExchange := s.cfg.Exchange + rabbitMQDeadLetterSuffix
	if err := ch.ExchangeDeclare(dlqExchange, "fanout", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("rabbitmq dead-letter exchange declare: %w", err)
	}

	queue, err := ch.QueueDeclare(s.cfg.Queue, true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("rabbitmq queue declare: %w", err)
	}
	if err := ch.ExchangeDeclare(s.cfg.Exchange, "fanout", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("rabbitmq exchange declare: %w", err)
	}
	if err := ch.QueueBind(queue.Name, "", s.cfg.Exchange, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("rabbitmq queue bind: %w", err)
	}
	if err := ch.Qos(1, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("rabbitmq qos: %w", err)
	}

	s.conn = conn
	s.channel = ch
	s.healthy.Store(true)
	return nil
}

// Subscribe registers an event handler for the configured RabbitMQ queue.
// It reconnects with backoff if the broker is temporarily unavailable.
func (s *rabbitMQSubscriber) Subscribe(ctx context.Context, topic string, handler ports.EventHandler) error {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		backoff := 200 * time.Millisecond
		for {
			if ctx.Err() != nil {
				s.healthy.Store(false)
				return
			}

			if err := s.connect(ctx); err != nil {
				s.healthy.Store(false)
				if s.log != nil {
					s.log.Warn(ctx, "rabbitmq reconnect failed", "error", err)
				}
				time.Sleep(backoff)
				if backoff < 2*time.Second {
					backoff *= 2
				}
				continue
			}

			backoff = 200 * time.Millisecond
			if s.log != nil {
				s.log.Info(ctx, "rabbitmq subscriber started", "queue", s.cfg.Queue)
			}

			deliveries, err := s.channel.Consume(
				s.cfg.Queue,
				"",
				false,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				s.healthy.Store(false)
				if s.log != nil {
					s.log.Warn(ctx, "rabbitmq consume failed", "error", err)
				}
				_ = s.closeLocked()
				time.Sleep(backoff)
				if backoff < 2*time.Second {
					backoff *= 2
				}
				continue
			}

			for {
				select {
				case <-ctx.Done():
					s.healthy.Store(false)
					return
				case delivery, ok := <-deliveries:
					if !ok {
						s.healthy.Store(false)
						if s.log != nil {
							s.log.Warn(ctx, "rabbitmq delivery channel closed")
							s.log.Info(ctx, "rabbitmq reconnecting", "queue", s.cfg.Queue)
						}
						_ = s.closeLocked()
						time.Sleep(backoff)
						if backoff < 2*time.Second {
							backoff *= 2
						}
						goto reconnect
					}

					event := amqpMsgToEvent(delivery)
					if err := s.handleDelivery(ctx, delivery, event, handler); err != nil && s.log != nil {
						s.log.Warn(ctx, "rabbitmq event processing failed", "error", err)
					}
				}
			}
		reconnect:
		}
	}()

	return nil
}

func (s *rabbitMQSubscriber) handleDelivery(
	ctx context.Context,
	delivery amqp.Delivery,
	event ports.Event,
	handler ports.EventHandler,
) error {
	for attempt := 1; attempt <= rabbitMQRetryAttempts; attempt++ {
		if err := handler(ctx, event); err == nil {
			if ackErr := delivery.Ack(false); ackErr != nil {
				return fmt.Errorf("rabbitmq ack: %w", ackErr)
			}
			return nil
		}
		if attempt < rabbitMQRetryAttempts {
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
			continue
		}

		return s.deadLetter(ctx, delivery, attempt)
	}
	return nil
}

func (s *rabbitMQSubscriber) deadLetter(ctx context.Context, delivery amqp.Delivery, attempt int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.channel == nil {
		return fmt.Errorf("rabbitmq dead-letter: channel unavailable")
	}

	dlqExchange := s.cfg.Exchange + rabbitMQDeadLetterSuffix
	headers := amqp.Table{}
	for k, v := range delivery.Headers {
		headers[k] = v
	}
	headers["retry_count"] = attempt

	if err := s.channel.PublishWithContext(ctx, dlqExchange, "", false, false, amqp.Publishing{
		ContentType:  delivery.ContentType,
		Body:         delivery.Body,
		DeliveryMode: amqp.Persistent,
		Headers:      headers,
		MessageId:    delivery.MessageId,
	}); err != nil {
		return fmt.Errorf("rabbitmq dead-letter publish: %w", err)
	}
	if err := delivery.Ack(false); err != nil {
		return fmt.Errorf("rabbitmq dead-letter ack: %w", err)
	}
	return nil
}

// Close gracefully shuts down the RabbitMQ subscriber.
func (s *rabbitMQSubscriber) Close() error {
	if s.log != nil {
		s.log.Info(context.Background(), "closing rabbitmq subscriber")
	}
	s.healthy.Store(false)
	_ = s.closeLocked()
	s.wg.Wait()
	return nil
}

func (s *rabbitMQSubscriber) closeLocked() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	if s.channel != nil {
		if closeErr := s.channel.Close(); closeErr != nil {
			err = closeErr
		}
		s.channel = nil
	}
	if s.conn != nil {
		if closeErr := s.conn.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		s.conn = nil
	}
	return err
}

func (s *rabbitMQSubscriber) Healthy() bool {
	return s.healthy.Load()
}

func dialRabbitMQ(ctx context.Context, cfg config.RabbitMQBrokerConfig, logger ports.Logger) (*amqp.Connection, *amqp.Channel, error) {
	if cfg.URL == "" {
		return nil, nil, fmt.Errorf("rabbitmq url is required")
	}

	var lastErr error
	backoff := 200 * time.Millisecond
	for attempt := 1; attempt <= rabbitMQRetryAttempts; attempt++ {
		conn, err := amqp.Dial(cfg.URL)
		if err != nil {
			lastErr = err
		} else {
			ch, err := conn.Channel()
			if err != nil {
				_ = conn.Close()
				lastErr = err
			} else {
				if err := ch.ExchangeDeclare(cfg.Exchange, "fanout", true, false, false, false, nil); err != nil {
					_ = ch.Close()
					_ = conn.Close()
					lastErr = err
				} else {
					if logger != nil {
						logger.Info(ctx, "rabbitmq connected", "exchange", cfg.Exchange, "attempt", attempt)
					}
					return conn, ch, nil
				}
			}
		}

		if logger != nil {
			logger.Warn(ctx, "rabbitmq connect failed", "attempt", attempt, "error", lastErr)
		}
		if attempt < rabbitMQRetryAttempts {
			time.Sleep(backoff)
			if backoff < 2*time.Second {
				backoff *= 2
			}
		}
	}

	return nil, nil, fmt.Errorf("rabbitmq dial: %w", lastErr)
}

// amqpMsgToEvent converts an amqp.Delivery to a ports.Event.
func amqpMsgToEvent(delivery amqp.Delivery) ports.Event {
	metadata := make(map[string]string, len(delivery.Headers))
	for k, v := range delivery.Headers {
		switch val := v.(type) {
		case string:
			metadata[k] = val
		case []byte:
			metadata[k] = string(val)
		default:
			metadata[k] = fmt.Sprint(val)
		}
	}
	metadata["event_id"] = delivery.MessageId
	metadata["content_type"] = delivery.ContentType
	return ports.Event{
		Topic:    delivery.Exchange,
		Key:      delivery.RoutingKey,
		Payload:  delivery.Body,
		Metadata: metadata,
	}
}
