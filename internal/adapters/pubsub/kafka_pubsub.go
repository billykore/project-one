package pubsub

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/billykore/project-one/internal/config"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/segmentio/kafka-go"
)

const kafkaDeadLetterSuffix = ".dlq"
const kafkaRetryAttempts = 3

// kafkaPublisher implements ports.Publisher using segmentio/kafka-go.
type kafkaPublisher struct {
	writer  *kafka.Writer
	log     ports.Logger
	healthy atomic.Bool
}

// NewKafkaPublisher creates a Kafka-backed ports.Publisher.
func NewKafkaPublisher(cfg config.KafkaBrokerConfig, logger ports.Logger) (ports.Publisher, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka: at least one broker is required")
	}
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	if logger != nil {
		logger.Info(context.Background(), "kafka publisher created", "brokers", cfg.Brokers)
	}
	p := &kafkaPublisher{writer: writer, log: logger}
	p.healthy.Store(true)
	return p, nil
}

// Publish sends an event to the Kafka topic specified in event.Topic.
func (p *kafkaPublisher) Publish(ctx context.Context, event ports.Event) error {
	msg := kafka.Message{
		Topic: event.Topic,
		Key:   []byte(event.Key),
		Value: event.Payload,
	}
	for k, v := range event.Metadata {
		msg.Headers = append(msg.Headers, kafka.Header{Key: k, Value: []byte(v)})
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.healthy.Store(false)
		if p.log != nil {
			p.log.Warn(ctx, "kafka publish failed", "topic", event.Topic, "error", err)
		}
		return fmt.Errorf("kafka publish: %w", err)
	}
	p.healthy.Store(true)
	if p.log != nil {
		p.log.Debug(ctx, "kafka message published", "topic", event.Topic, "key", event.Key)
	}
	return nil
}

// Close gracefully shuts down the Kafka writer, flushing any pending messages.
func (p *kafkaPublisher) Close() error {
	if p.log != nil {
		p.log.Info(context.Background(), "closing kafka publisher")
	}
	p.healthy.Store(false)
	return p.writer.Close()
}

func (p *kafkaPublisher) Healthy() bool {
	return p.healthy.Load()
}

// kafkaSubscriber implements ports.Subscriber using segmentio/kafka-go.
type kafkaSubscriber struct {
	brokers     []string
	topicPrefix string
	groupID     string
	log         ports.Logger
	healthy     atomic.Bool
	wg          sync.WaitGroup
}

// NewKafkaSubscriber creates a Kafka-backed ports.Subscriber.
func NewKafkaSubscriber(cfg config.KafkaBrokerConfig, logger ports.Logger) (ports.Subscriber, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka: at least one broker is required")
	}
	if logger != nil {
		logger.Info(context.Background(), "kafka subscriber created", "brokers", cfg.Brokers, "group", cfg.ConsumerGroup)
	}
	s := &kafkaSubscriber{
		brokers:     cfg.Brokers,
		topicPrefix: cfg.TopicPrefix,
		groupID:     cfg.ConsumerGroup,
		log:         logger,
	}
	s.healthy.Store(false)
	return s, nil
}

// Subscribe registers an event handler for the given Kafka topic.
// It polls messages sequentially from the broker and retries each event locally
// before dead-lettering it to a topic suffix.
func (s *kafkaSubscriber) Subscribe(ctx context.Context, topic string, handler ports.EventHandler) error {
	topicName := topic
	if s.topicPrefix != "" {
		topicName = s.topicPrefix + "." + topic
	}
	dlqTopic := topicName + kafkaDeadLetterSuffix

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  s.brokers,
		Topic:    topicName,
		GroupID:  s.groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	dlqWriter := &kafka.Writer{
		Addr:         kafka.TCP(s.brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer func() {
			_ = reader.Close()
			_ = dlqWriter.Close()
			s.healthy.Store(false)
			if s.log != nil {
				s.log.Info(context.Background(), "kafka subscriber stopped", "topic", topicName)
			}
		}()

		s.healthy.Store(true)
		if s.log != nil {
			s.log.Info(ctx, "kafka subscriber started", "topic", topicName)
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				s.healthy.Store(false)
				if s.log != nil {
					s.log.Warn(ctx, "kafka read error", "topic", topicName, "error", err)
					s.log.Info(ctx, "kafka reconnecting", "topic", topicName)
				}
				time.Sleep(200 * time.Millisecond)
				continue
			}

			event := kafkaMsgToEvent(msg)
			if err := s.handleKafkaEvent(ctx, reader, dlqWriter, dlqTopic, msg, event, handler); err != nil && s.log != nil {
				s.log.Warn(ctx, "kafka event processing failed", "topic", topicName, "error", err)
			}
		}
	}()

	return nil
}

func (s *kafkaSubscriber) handleKafkaEvent(
	ctx context.Context,
	reader *kafka.Reader,
	dlqWriter *kafka.Writer,
	dlqTopic string,
	msg kafka.Message,
	event ports.Event,
	handler ports.EventHandler,
) error {
	for attempt := 1; attempt <= kafkaRetryAttempts; attempt++ {
		if err := handler(ctx, event); err == nil {
			return reader.CommitMessages(ctx, msg)
		}
		if attempt < kafkaRetryAttempts {
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
			continue
		}

		dlqEvent := msg
		dlqEvent.Topic = dlqTopic
		dlqEvent.Headers = append(dlqEvent.Headers, kafka.Header{Key: "retry_count", Value: []byte(fmt.Sprintf("%d", attempt))})
		if err := dlqWriter.WriteMessages(ctx, dlqEvent); err != nil {
			return fmt.Errorf("kafka dead-letter publish: %w", err)
		}
		return reader.CommitMessages(ctx, msg)
	}
	return nil
}

// Close gracefully shuts down the Kafka subscriber.
func (s *kafkaSubscriber) Close() error {
	if s.log != nil {
		s.log.Info(context.Background(), "kafka subscriber close requested")
	}
	s.healthy.Store(false)
	s.wg.Wait()
	return nil
}

func (s *kafkaSubscriber) Healthy() bool {
	return s.healthy.Load()
}

// kafkaMsgToEvent converts a kafka.Message to a ports.Event.
func kafkaMsgToEvent(msg kafka.Message) ports.Event {
	metadata := make(map[string]string, len(msg.Headers))
	for _, h := range msg.Headers {
		metadata[h.Key] = string(h.Value)
	}
	return ports.Event{
		Topic:    msg.Topic,
		Key:      string(msg.Key),
		Payload:  msg.Value,
		Metadata: metadata,
	}
}
