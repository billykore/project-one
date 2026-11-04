// Package outbox provides durable, retryable broker delivery for technical events.
package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/billykore/project-one/internal/pkg/database"
	"github.com/billykore/project-one/internal/pkg/ports"
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
	FailedAt    *time.Time
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

func (p *Publisher) Publish(ctx context.Context, event ports.Event) error {
	return Enqueue(ctx, database.FromContext(ctx, p.db), Message{EventID: event.Metadata["event_id"], Topic: event.Topic, Key: event.Key, Payload: event.Payload, Metadata: event.Metadata})
}

func (p *Publisher) Close() error { return nil }

func (p *Publisher) InTransaction(ctx context.Context, fn func(context.Context) error) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(database.WithTransaction(ctx, tx))
	})
}

const (
	// DefaultMaxAttempts bounds delivery attempts. With the retry backoff capped
	// at 64 seconds, 60 attempts ride out roughly an hour of broker outage.
	DefaultMaxAttempts = 60
	// DefaultRetention is how long published messages are kept before cleanup.
	DefaultRetention = 7 * 24 * time.Hour
	// DefaultCleanupEvery is how often the dispatcher deletes expired messages.
	DefaultCleanupEvery = time.Hour

	cleanupBatchSize = 1000
	maxBackoffShift  = 6
)

// Option customizes a Dispatcher.
type Option func(*Dispatcher)

// WithMaxAttempts sets how many failed publications park a message as failed.
// Values below one keep the default.
func WithMaxAttempts(attempts int) Option {
	return func(d *Dispatcher) {
		if attempts > 0 {
			d.maxAttempts = attempts
		}
	}
}

// WithRetention sets how long published messages are kept. Values below or
// equal to zero keep the default.
func WithRetention(retention time.Duration) Option {
	return func(d *Dispatcher) {
		if retention > 0 {
			d.retention = retention
		}
	}
}

// WithCleanupEvery sets how often expired published messages are deleted.
// Values below or equal to zero keep the default.
func WithCleanupEvery(every time.Duration) Option {
	return func(d *Dispatcher) {
		if every > 0 {
			d.cleanupEvery = every
		}
	}
}

// Dispatcher publishes pending messages and records completion or a retry delay.
// Duplicate delivery is possible if a process stops after broker publication and
// before marking a message published; consumers must use EventID idempotently.
// A message that fails maxAttempts publications is parked with failed_at set
// and no longer retried; published messages are deleted after the retention.
type Dispatcher struct {
	db           *gorm.DB
	publisher    ports.Publisher
	log          ports.Logger
	pollEvery    time.Duration
	batchSize    int
	lease        time.Duration
	maxAttempts  int
	retention    time.Duration
	cleanupEvery time.Duration
}

func NewDispatcher(db *gorm.DB, publisher ports.Publisher, log ports.Logger, pollEvery time.Duration, batchSize int, opts ...Option) *Dispatcher {
	if pollEvery <= 0 {
		pollEvery = time.Second
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	d := &Dispatcher{
		db:           db,
		publisher:    publisher,
		log:          log,
		pollEvery:    pollEvery,
		batchSize:    batchSize,
		lease:        30 * time.Second,
		maxAttempts:  DefaultMaxAttempts,
		retention:    DefaultRetention,
		cleanupEvery: DefaultCleanupEvery,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// Start runs until ctx is cancelled.
func (d *Dispatcher) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(d.pollEvery)
		defer ticker.Stop()
		var lastCleanup time.Time
		for {
			if err := d.DispatchOnce(ctx); err != nil {
				d.log.Warn(ctx, "outbox dispatch cycle failed", "error", err)
			}
			if time.Since(lastCleanup) >= d.cleanupEvery {
				lastCleanup = time.Now()
				if deleted, err := d.Cleanup(ctx); err != nil {
					d.log.Warn(ctx, "outbox cleanup failed", "error", err)
				} else if deleted > 0 {
					d.log.Info(ctx, "outbox cleanup deleted published messages", "count", deleted)
				}
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
		err := d.publisher.Publish(ctx, ports.Event{Topic: message.Topic, Key: message.Key, Payload: message.Payload, Metadata: message.Metadata})
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
			Where("published_at IS NULL AND failed_at IS NULL AND available_at <= ? AND (locked_until IS NULL OR locked_until < ?)", now, now).
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
	now := time.Now().UTC()
	updates := map[string]any{
		"attempts":     gorm.Expr("attempts + 1"),
		"locked_until": nil,
	}
	failed := message.Attempts+1 >= d.maxAttempts
	if failed {
		updates["failed_at"] = now
	} else {
		// Exponential backoff is bounded so a transient outage cannot spin but
		// also cannot hide an event indefinitely.
		updates["available_at"] = now.Add(time.Second << min(message.Attempts, maxBackoffShift))
	}
	if err := d.db.WithContext(ctx).Model(&messageModel{}).Where("id = ?", message.ID).Updates(updates).Error; err != nil {
		return fmt.Errorf("release outbox message: %w", err)
	}
	if failed {
		d.log.Error(ctx, "outbox message exhausted delivery attempts; parked as failed", "event_id", message.EventID, "attempts", message.Attempts+1)
	}
	return nil
}

// Cleanup deletes published messages older than the retention, in bounded
// batches so a large backlog never holds one long transaction. Failed messages
// are kept for an operator. It is exported for deterministic tests.
func (d *Dispatcher) Cleanup(ctx context.Context) (int64, error) {
	cutoff := time.Now().UTC().Add(-d.retention)
	var total int64
	for {
		result := d.db.WithContext(ctx).Exec(
			`DELETE FROM notification_outbox WHERE id IN (
				SELECT id FROM notification_outbox
				WHERE published_at IS NOT NULL AND published_at < ?
				ORDER BY published_at LIMIT ?)`,
			cutoff, cleanupBatchSize,
		)
		if result.Error != nil {
			return total, fmt.Errorf("delete published outbox messages: %w", result.Error)
		}
		total += result.RowsAffected
		// A cancelled context fails the next batch's Exec, which ends the loop.
		if result.RowsAffected < cleanupBatchSize {
			return total, nil
		}
	}
}

// Stats summarizes undelivered messages for monitoring.
type Stats struct {
	// Pending counts messages that are still being delivered or retried.
	Pending int64
	// OldestPendingAge is how long the oldest pending message has waited, or
	// zero when nothing is pending.
	OldestPendingAge time.Duration
	// Failed counts messages parked after exhausting their delivery attempts.
	Failed int64
}

// StatsReader reads outbox backlog statistics.
type StatsReader struct{ db *gorm.DB }

func NewStatsReader(db *gorm.DB) *StatsReader { return &StatsReader{db: db} }

// Stats returns the current backlog in a single query. The age is computed by
// the database clock so it is not skewed by the application host's clock.
func (r *StatsReader) Stats(ctx context.Context) (Stats, error) {
	var row struct {
		Pending          int64
		Failed           int64
		OldestAgeSeconds float64
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) FILTER (WHERE failed_at IS NULL) AS pending,
			COUNT(*) FILTER (WHERE failed_at IS NOT NULL) AS failed,
			COALESCE(EXTRACT(EPOCH FROM now() - MIN(created_at) FILTER (WHERE failed_at IS NULL)), 0)::float8 AS oldest_age_seconds
		FROM notification_outbox
		WHERE published_at IS NULL`).Scan(&row).Error
	if err != nil {
		return Stats{}, fmt.Errorf("read outbox stats: %w", err)
	}
	return Stats{
		Pending:          row.Pending,
		OldestPendingAge: time.Duration(row.OldestAgeSeconds * float64(time.Second)),
		Failed:           row.Failed,
	}, nil
}
