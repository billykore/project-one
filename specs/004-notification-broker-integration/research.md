# Research: Notification Broker Integration

**Date**: 2026-07-25 | **Feature**: `004-notification-broker-integration`

## 1. Go Kafka Client Selection

### Decision: `segmentio/kafka-go`

### Rationale

- **Pure Go** — no CGO dependency, no librdkafka compilation. Consistent with the project's "ponytail" philosophy of minimal dependencies.
- **Simple API** — `kafka.Writer` maps cleanly to `ports.Publisher` (Publish), `kafka.Reader` maps to `ports.Subscriber` (Subscribe via polling loop). No channel-based callback needed — the `EventHandler` pattern from `ports.Subscriber` translates naturally.
- **Context-aware** — `WriteMessages(ctx, ...)` and `ReadMessage(ctx)` both accept `context.Context`, aligning with the existing port signatures.
- **No external runtime** — unlike `confluent-kafka-go` which wraps a C library, `segmentio/kafka-go` compiles everywhere Go compiles (including CI, Docker scratch images).
- **Active maintenance** — widely used, 12k+ GitHub stars, regular releases.
- **Built-in reconnection** — `kafka.Writer` and `kafka.Reader` handle broker reconnection internally with configurable retry settings.

### Alternatives Considered

| Library | Why Rejected |
|---------|-------------|
| `confluent-kafka-go` | Requires CGO + librdkafka; complicates builds and Docker images. |
| `IBM/sarama` | More complex API, lower-level; overkill for notification use case. |
| `twmb/franz-go` | Good pure-Go option but newer ecosystem, smaller community. |

---

## 2. Go RabbitMQ Client Selection

### Decision: `rabbitmq/amqp091-go`

### Rationale

- **Official library** — maintained by the RabbitMQ team.
- **Stable API** — the `amqp.Channel` provides `Publish` and `Consume` methods that map cleanly to `ports.Publisher` and `ports.Subscriber`.
- **Connection resilience** — supports `NotifyClose` and `NotifyBlocked` channels for detecting connection failures and triggering reconnection.
- **Exchanges + Queues** — RabbitMQ's exchange/queue model maps well: notifications go to a fanout exchange, with durable queues bound for each consumer type.

### Alternatives Considered

| Library | Why Rejected |
|---------|-------------|
| `streadway/amqp` | Deprecated; superseded by `rabbitmq/amqp091-go`. |
| `wagslane/go-rabbitmq` | Opinionated wrapper that hides connection details; harder to implement custom reconnection logic. |

---

## 3. Interface Compatibility Analysis

### `ports.Publisher` → Kafka

```go
// kafka.Writer maps cleanly:
type kafkaPublisher struct {
    writer *kafka.Writer
}

func (p *kafkaPublisher) Publish(ctx context.Context, event ports.Event) error {
    msg := kafka.Message{
        Topic:   event.Topic,
        Key:     []byte(event.Key),
        Value:   event.Payload,
        Headers: toKafkaHeaders(event.Metadata),
    }
    return p.writer.WriteMessages(ctx, msg)
}

func (p *kafkaPublisher) Close() error {
    return p.writer.Close()
}
```

### `ports.Publisher` → RabbitMQ

```go
// amqp.Channel with confirm mode maps well:
type rabbitPublisher struct {
    conn    *amqp.Connection
    channel *amqp.Channel
}

func (p *rabbitPublisher) Publish(ctx context.Context, event ports.Event) error {
    return p.channel.PublishWithContext(ctx,
        exchangeFromTopic(event.Topic), // configurable mapping
        event.Key,                       // routing key
        false, false,                    // mandatory, immediate
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         event.Payload,
            DeliveryMode: amqp.Persistent,
            Headers:      toAmqpTable(event.Metadata),
        },
    )
}
```

### `ports.Subscriber` → Kafka

```go
func (s *kafkaSubscriber) Subscribe(ctx context.Context, topic string, handler ports.EventHandler) error {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: s.brokers,
        Topic:   topic,
        GroupID: s.groupID,
    })
    go func() {
        defer reader.Close()
        for {
            msg, err := reader.ReadMessage(ctx)
            if err != nil {
                if errors.Is(err, context.Canceled) { return }
                s.log.Warn(...)
                continue
            }
            event := kafkaMsgToEvent(msg)
            if err := handler(ctx, event); err != nil {
                s.log.Error(...)
                // retry handled by Kafka consumer group offset
            }
        }
    }()
    return nil
}
```

### `ports.Subscriber` → RabbitMQ

```go
func (s *rabbitSubscriber) Subscribe(ctx context.Context, topic string, handler ports.EventHandler) error {
    deliveries, err := s.channel.ConsumeWithContext(ctx, queueName, consumerTag, ...)
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            case delivery, ok := <-deliveries:
                if !ok { return }
                event := deliveryToEvent(delivery)
                if err := handler(ctx, event); err != nil {
                    delivery.Nack(false, true) // requeue
                } else {
                    delivery.Ack(false)
                }
            }
        }
    }()
    return nil
}
```

### Verdict: Both Kafka and RabbitMQ cleanly satisfy `ports.Publisher` and `ports.Subscriber`.

---

## 4. Connection Management & Resilience Patterns

### Kafka

- **Connection**: `kafka.Writer` and `kafka.Reader` handle broker discovery and reconnection internally. The transport layer auto-reconnects.
- **Publish timeout**: Set `WriteTimeout` on the writer (e.g., 5s). On timeout, log error and return — callers already treat publish as best-effort.
- **Consumer restart**: On fatal consumer errors, the outer `Listen` method returns an error. `main.go` currently calls `app.notificationHandler.Listen(ctx)` in a goroutine — it Fatal-logs on error. For broker resilience, change to a retry-with-backoff wrapper.
- **Graceful shutdown**: `kafka.Writer.Close()` flushes pending writes; `kafka.Reader.Close()` commits current offset.

### RabbitMQ

- **Connection**: Must implement explicit reconnection. Use `amqp.Connection.NotifyClose` channel to detect disconnects. On close, reconnect with capped exponential backoff.
- **Channel recovery**: After reconnection, re-declare exchanges/queues and rebind consumers. RabbitMQ channels don't auto-recover across connections.
- **Publisher confirms**: Enable confirm mode for at-least-once delivery. Wait for confirmation or timeout on each publish.
- **Graceful shutdown**: `Channel.Close()` then `Connection.Close()`; wait for in-flight confirms.

### Shared Pattern: Capped Exponential Backoff

```go
func withBackoff(maxAttempts int, base, max time.Duration, fn func() error) error {
    for attempt := 0; attempt < maxAttempts; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }
        if attempt == maxAttempts-1 {
            return err
        }
        backoff := min(base * time.Duration(1<<attempt), max)
        time.Sleep(backoff)
    }
    return nil
}
```

---

## 5. Idempotency Strategy

### Decision: Database UNIQUE constraint on event ID

### Implementation

- Add `EventID string` field to `domain.Notification` entity.
- Add migration: `ALTER TABLE notifications ADD COLUMN event_id VARCHAR(64) UNIQUE`.
- The consumer (in `notification_handler.go`) sets `notification.EventID = event.Metadata["event_id"]` before calling `SaveNotification`.
- The repository's `Create` method attempts INSERT. On PostgreSQL unique violation (`23505`), the consumer treats it as a duplicate and ACKs the message without error.
- Kafka: event ID derived from `{topic}-{partition}-{offset}` or a UUID set by the publisher.
- RabbitMQ: event ID from `message_id` header or a UUID set by the publisher.

### Why not alternatives

| Alternative | Why Rejected |
|-------------|--------------|
| In-memory dedup cache | Lost on restart; doesn't survive across consumer instances. |
| SELECT-before-INSERT | Race condition between check and insert; requires additional locking. |
| Broker exactly-once | Ties implementation to specific broker features; Kafka exactly-once is complex. |

---

## 6. Schema Versioning

### Decision: Embed schema version in the event metadata

Each notification event carries a `schema_version` metadata header (e.g., `"1.0"`). Consumers check the version and deserialize accordingly. The schema is defined in the `contracts/` directory as a JSON Schema document.

For v1.0, the schema is backward-compatible with the existing `domain.Notification` JSON structure, adding only `event_id` and `schema_version` fields.

---

## 7. Environment Selection Strategy

### Decision: Env-based broker selection with in-memory fallback

```yaml
# config.yaml
message_broker:
  type: "kafka"        # "kafka", "rabbitmq", "inmemory"
  kafka:
    brokers: ["localhost:9092"]
    topic_prefix: "project1"
    consumer_group: "notification-consumer"
  rabbitmq:
    url: "amqp://guest:guest@localhost:5672/"
    exchange: "project1.notifications"
    queue: "notifications"
```

- `type: "inmemory"` — uses existing `inMemoryPubSub` (default for local dev, no external dependency).
- `type: "kafka"` — uses the Kafka adapter.
- `type: "rabbitmq"` — uses the RabbitMQ adapter.
- Factory function in `cmd/main.go` selects the appropriate implementation at startup.

---

## 8. Anti-Patterns Avoided

| Anti-Pattern | Mitigation |
|-------------|-----------|
| Blocking HTTP handlers on broker publish | Publish calls use context with timeout (≤2s); failures logged, not propagated. |
| Infinite retry loops | Capped exponential backoff with max attempts; dead-letter after exhaustion. |
| Broker-specific code in use cases | `ports.Publisher`/`ports.Subscriber` interfaces remain broker-agnostic. |
| Hardcoded credentials | All credentials from `config.yaml` or env vars via viper. |
| No graceful shutdown | `Publisher.Close()` flushes; `Subscriber.Close()` finishes in-flight + commits offsets. |
