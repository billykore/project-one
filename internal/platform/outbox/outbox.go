// Package outbox provides durable, retryable broker delivery for technical events.
package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/billykore/project-one/internal/platform/database"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Message is an event that must be published after its enclosing transaction commits.
type Message struct {
	ID          int64
	EventID     string
	Topic       string
	Key         string
	Payload     []byte
	Metadata    map[string]string
	Attempts    int
	AvailableAt time.Time
}

type messageModel struct {
	ID          int64 `gorm:"primaryKey"`
	EventID     string
	Topic       string
	MessageKey  string
	Payload     []byte
	Metadata    string `gorm:"type:jsonb"`
	Attempts    int
	AvailableAt time.Time
	LockedUntil *time.Time
	PublishedAt *time.Time
	CreatedAt   time.Time
}

func (messageModel) TableName() string { return "notification_outbox" }

// Enqueue persists a message using the supplied transaction. The caller owns
// the transaction so the domain mutation and its event commit or roll back together.
func Enqueue(ctx context.Context, tx *gorm.DB, message Message) error {
	metadata, err := json.Marshal(message.Metadata)
	if err != nil {
		return fmt.Errorf("marshal outbox metadata: %w", err)
	}
	model := messageModel{
		EventID:    message.EventID,
		Topic:      message.Topic,
		MessageKey: message.Key,
		Payload:    message.Payload,
		Metadata:   string(metadata),
	}
	if err := tx.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("insert outbox message: %w", err)
	}
	return nil
}

// Publisher persists messages in the outbox. Dispatcher is responsible for
// forwarding them to the configured broker after their transaction commits.
type Publisher struct{ db *gorm.DB }

func NewPublisher(db *gorm.DB) *Publisher { return &Publisher{db: db} }

func (p *Publisher) Publish(ctx context.Context, event platformports.Event) error {
	return Enqueue(ctx, database.FromContext(ctx, p.db), Message{EventID: event.Metadata["event_id"], Topic: event.Topic, Key: event.Key, Payload: event.Payload, Metadata: event.Metadata})
}

func (p *Publisher) Close() error { return nil }

func (p *Publisher) InTransaction(ctx context.Context, fn func(context.Context) error) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(database.WithTransaction(ctx, tx))
	})
}

// Dispatcher publishes pending messages and records completion or a retry delay.
// Duplicate delivery is possible if a process stops after broker publication and
// before marking a message published; consumers must use EventID idempotently.
type Dispatcher struct {
	db        *gorm.DB
	publisher platformports.Publisher
	log       platformports.Logger
	pollEvery time.Duration
	batchSize int
	lease     time.Duration
}

func NewDispatcher(db *gorm.DB, publisher platformports.Publisher, log platformports.Logger, pollEvery time.Duration, batchSize int) *Dispatcher {
	if pollEvery <= 0 {
		pollEvery = time.Second
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	return &Dispatcher{db: db, publisher: publisher, log: log, pollEvery: pollEvery, batchSize: batchSize, lease: 30 * time.Second}
}

// Start runs until ctx is cancelled.
func (d *Dispatcher) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(d.pollEvery)
		defer ticker.Stop()
		for {
			if err := d.DispatchOnce(ctx); err != nil {
				d.log.Warn(ctx, "outbox dispatch cycle failed", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// DispatchOnce claims and attempts a bounded batch. It is exported for deterministic tests.
func (d *Dispatcher) DispatchOnce(ctx context.Context) error {
	messages, err := d.claim(ctx)
	if err != nil {
		return err
	}
	for _, message := range messages {
		err := d.publisher.Publish(ctx, platformports.Event{Topic: message.Topic, Key: message.Key, Payload: message.Payload, Metadata: message.Metadata})
		if err != nil {
			d.log.Warn(ctx, "outbox publish failed; event will be retried", "event_id", message.EventID, "error", err)
			if releaseErr := d.release(ctx, message); releaseErr != nil {
				return releaseErr
			}
			continue
		}
		if err := d.markPublished(ctx, message.ID); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) claim(ctx context.Context) ([]Message, error) {
	now := time.Now().UTC()
	leaseUntil := now.Add(d.lease)
	var models []messageModel
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("published_at IS NULL AND available_at <= ? AND (locked_until IS NULL OR locked_until < ?)", now, now).
			Order("id ASC").Limit(d.batchSize).Find(&models).Error; err != nil {
			return err
		}
		if len(models) == 0 {
			return nil
		}
		ids := make([]int64, 0, len(models))
		for _, model := range models {
			ids = append(ids, model.ID)
		}
		return tx.Model(&messageModel{}).Where("id IN ?", ids).Update("locked_until", leaseUntil).Error
	})
	if err != nil {
		return nil, fmt.Errorf("claim outbox messages: %w", err)
	}
	messages := make([]Message, 0, len(models))
	for _, model := range models {
		metadata := map[string]string{}
		if err := json.Unmarshal([]byte(model.Metadata), &metadata); err != nil {
			return nil, fmt.Errorf("decode outbox metadata: %w", err)
		}
		messages = append(messages, Message{ID: model.ID, EventID: model.EventID, Topic: model.Topic, Key: model.MessageKey, Payload: model.Payload, Metadata: metadata, Attempts: model.Attempts, AvailableAt: model.AvailableAt})
	}
	return messages, nil
}

func (d *Dispatcher) markPublished(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	if err := d.db.WithContext(ctx).Model(&messageModel{}).Where("id = ?", id).Updates(map[string]any{"published_at": now, "locked_until": nil}).Error; err != nil {
		return fmt.Errorf("mark outbox message published: %w", err)
	}
	return nil
}

func (d *Dispatcher) release(ctx context.Context, message Message) error {
	// Exponential backoff is bounded so a transient outage cannot spin but also
	// cannot hide an event indefinitely.
	delay := time.Second << min(message.Attempts, 6)
	if err := d.db.WithContext(ctx).Model(&messageModel{}).Where("id = ?", message.ID).Updates(map[string]any{
		"attempts":     gorm.Expr("attempts + 1"),
		"available_at": time.Now().UTC().Add(delay),
		"locked_until": nil,
	}).Error; err != nil {
		return fmt.Errorf("release outbox message: %w", err)
	}
	return nil
}
