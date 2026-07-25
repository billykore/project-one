# Quickstart: Notification Broker Integration

**Date**: 2026-07-25 | **Feature**: `004-notification-broker-integration`

## Prerequisites

- Go 1.26+
- PostgreSQL (existing project database)
- One of: Docker (for local Kafka/RabbitMQ) OR a running broker instance
- Project backend running per `README.md` instructions

## Setup — In-Memory (Default, No External Broker)

The in-memory pubsub is the default. No additional setup needed.

```bash
# config.yaml — no message_broker section or type: inmemory
make run
```

## Setup — Kafka (Local via Docker)

```bash
# Start Kafka + Zookeeper
docker run -d --name kafka-zookeeper -p 2181:2181 zookeeper:3.9
docker run -d --name kafka-broker -p 9092:9092 \
  -e KAFKA_ZOOKEEPER_CONNECT=host.docker.internal:2181 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
  confluentinc/cp-kafka:7.6.0

# Add to config.yaml:
# message_broker:
#   type: kafka
#   kafka:
#     brokers: ["localhost:9092"]
#     topic_prefix: "project1"
#     consumer_group: "notification-consumer"

make run
```

## Setup — RabbitMQ (Local via Docker)

```bash
# Start RabbitMQ with management UI
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 \
  rabbitmq:3.13-management

# Add to config.yaml:
# message_broker:
#   type: rabbitmq
#   rabbitmq:
#     url: "amqp://guest:guest@localhost:5672/"
#     exchange: "project1.notifications"
#     queue: "notifications"

make run
```

## Validation Scenarios

### Scenario 1: Publish and Consume a Follow Notification

**Goal**: Verify a follow notification is published to the broker and consumed, persisted, and delivered.

**Steps**:
1. Start the backend with a message broker configured.
2. Register two users: `alice` and `bob`.
3. Login as `alice`. Follow `bob`:
   ```bash
   curl -X POST http://localhost:8080/users/bob/followers \
     -H "Authorization: Bearer <alice_token>"
   ```
4. Check the broker (Kafka topic `project1.notifications` / RabbitMQ queue `notifications`): verify a message exists with `type: "follow"`.
5. Login as `bob`. Check notifications:
   ```bash
   curl http://localhost:8080/notifications \
     -H "Authorization: Bearer <bob_token>"
   ```
6. Verify the response includes a follow notification from `alice`.

**Expected Outcome**: Notification appears in `GET /notifications` for `bob`.

### Scenario 2: Idempotency — Duplicate Event Handling

**Goal**: Verify that consuming the same event twice does not create a duplicate notification.

**Steps**:
1. Publish a follow notification (Scenario 1).
2. Verify `bob` has exactly 1 notification in `GET /notifications`.
3. Publish the same event again manually (via broker CLI or admin UI) with the same `event_id`.
4. Check `bob`'s notifications again — count must remain unchanged.
5. Check backend logs for "duplicate event skipped" or similar at DEBUG level.

**Expected Outcome**: Notification count unchanged; duplicate skipped.

### Scenario 3: Broker Unavailability Does Not Block User Operations

**Goal**: Verify that when the broker is down, user actions still succeed.

**Steps**:
1. Start the backend with a broker configured.
2. Stop the broker (Docker stop or kill the process).
3. Login and like a post:
   ```bash
   curl -X POST http://localhost:8080/posts/1/likes \
     -H "Authorization: Bearer <token>"
   ```
4. Verify the like succeeds (200 response, like count incremented).
5. Check backend logs: "failed to publish notification" warning logged, no panic or crash.
6. Restart the broker.
7. Like another post — verify notification is published and delivered normally.

**Expected Outcome**: Like succeeds regardless of broker state; broker recovery is automatic.

### Scenario 4: Graceful Shutdown — No Lost In-Flight Messages

**Goal**: Verify that messages published just before shutdown are flushed.

**Steps**:
1. Start the backend with broker configured.
2. Send a burst of 10 follow notifications rapidly.
3. Immediately send SIGTERM to the backend (`kill -TERM <pid>` or Ctrl+C).
4. Wait for shutdown to complete.
5. Restart the backend and broker.
6. Check recipient notifications — all 10 must be present.

**Expected Outcome**: Zero message loss during graceful shutdown.

### Scenario 5: Dead Letter for Malformed Events

**Goal**: Verify that unprocessable messages are routed to dead-letter after max retries.

**Steps**:
1. Publish a malformed message to the notification topic/queue (invalid JSON, missing required fields).
2. Check broker dead-letter queue/topic after the configured retry count.
3. Verify the malformed message appears in the dead-letter destination.
4. Verify the consumer continues processing valid messages normally.

**Expected Outcome**: Malformed messages dead-lettered; valid messages unaffected.

## Run Commands Summary

```bash
# Run tests
make test

# Run specific broker adapter tests
go test ./internal/adapters/pubsub/... -v -run "Kafka"
go test ./internal/adapters/pubsub/... -v -run "RabbitMQ"
go test ./internal/adapters/pubsub/... -v -run "InMemory"

# Generate mocks after interface changes
make mock

# Run lint + vet + test
make check
```

## External Consumer Setup

### Architecture

The message broker approach enables external consumers to subscribe to notification events independently of the backend. The backend publishes events to the broker; any consumer with access to the broker can subscribe to the notification topic/queue and process events.

```
┌──────────────┐    publish     ┌──────────────┐    subscribe    ┌──────────────────┐
│  Go Backend   │ ──────────▶  │   Broker      │ ◀───────────── │  External Consumer│
│ (use cases)   │              │(Kafka/RabbitMQ)│               │  (e.g., web app)  │
└──────────────┘              └──────────────┘               └──────────────────┘
```

### Kafka Consumer Example

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type NotificationEvent struct {
	EventID   string `json:"event_id"`
	Type      string `json:"type"`
	UserID    int    `json:"user_id"`
	ActorID   int    `json:"actor_id"`
	ActorName string `json:"actor_username"`
	PostID    int    `json:"post_id,omitempty"`
}

func main() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "project1.notifications",
		GroupID: "external-consumer",
	})
	defer reader.Close()

	ctx := context.Background()
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}
		var event NotificationEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("unmarshal error: %v", err)
			continue
		}
		fmt.Printf("Received: %s by %s for user %d\n", event.Type, event.ActorName, event.UserID)
	}
}
```

### RabbitMQ Consumer Example

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type NotificationEvent struct {
	EventID   string `json:"event_id"`
	Type      string `json:"type"`
	UserID    int    `json:"user_id"`
	ActorID   int    `json:"actor_id"`
	ActorName string `json:"actor_username"`
	PostID    int    `json:"post_id,omitempty"`
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	deliveries, err := ch.Consume("notifications", "", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	for d := range deliveries {
		var event NotificationEvent
		if err := json.Unmarshal(d.Body, &event); err != nil {
			log.Printf("unmarshal error: %v", err)
			continue
		}
		fmt.Printf("Received: %s by %s for user %d\n", event.Type, event.ActorName, event.UserID)
	}
}
```

### Health Check

The backend exposes a broker health check endpoint:

```bash
curl http://localhost:8080/health/broker
# {"type":"kafka","status":"connected","healthy":true}
```

### Broker Access Configuration

For external consumers to access the broker, the broker must be network-reachable and configured with appropriate authentication. See broker-specific documentation:

- **Kafka**: Configure `listeners`, `advertised.listeners`, and SASL/SSL in `server.properties`
- **RabbitMQ**: Configure `tcp_listeners`, authentication, and TLS in `rabbitmq.conf`

The backend does not manage broker infrastructure — these settings are environment-level configurations.
