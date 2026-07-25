# Implementation Plan: Notification Broker Integration

**Branch**: `004-notification-broker-integration` | **Date**: 2026-07-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/004-notification-broker-integration/spec.md`

## Summary

Replace the current in-memory pubsub with pluggable message broker adapters (Kafka and RabbitMQ), both satisfying the existing `ports.Publisher` and `ports.Subscriber` interfaces. Add event ID-based idempotency via a database unique constraint, standardized event schema with versioning, and environment-based broker selection with in-memory fallback for local development. The implementation adds two new adapter packages with ready-to-use code, a config-driven factory, and a migration for the idempotency column.

## Technical Context

**Language/Version**: Go 1.26+

**Primary Dependencies**: `segmentio/kafka-go` (Kafka adapter), `rabbitmq/amqp091-go` (RabbitMQ adapter), existing Echo + GORM + Viper stack

**Storage**: PostgreSQL (existing `notifications` table + 1 new column migration)

**Testing**: `go test` (standard library), GoMock for regenerated mocks, Testcontainers for broker integration tests (optional, Docker-dependent)

**Target Platform**: Linux server (Docker deployment via `deployments/docker-compose.yml`)

**Project Type**: Web service backend (Go API server)

**Performance Goals**: ≤2s p95 notification delivery latency; ≥1,000 notification events/min; broker publish timeout ≤2s; no blocking of HTTP handlers

**Constraints**: Must follow Clean Architecture (ports remain broker-agnostic); publish failures must not fail user-facing operations; must support graceful shutdown with message flush; config-driven broker selection at startup

**Scale/Scope**: 3 notification types (follow, like, comment); 2 broker adapters (Kafka, RabbitMQ) + 1 existing in-memory fallback; 1 new domain field; 1 new config struct; 2 new adapter packages

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Code Quality & Clean Architecture** | ✅ PASS | New broker adapters live in `internal/adapters/pubsub/`, implementing `ports.Publisher` and `ports.Subscriber`. Use cases remain unchanged — they only depend on the port interfaces. |
| **II. Testing Standards** | ✅ PASS | Adapter tests planned with GoMock mocks of the broker client interfaces. In-memory pubsub tests exist already and continue working. Idempotency behavior tested via notification use case unit tests. `make mock` regenerated after interface changes. |
| **III. User Experience Consistency** | ✅ PASS | No frontend changes in this feature. User-visible behavior unchanged — notifications still appear via WebSocket and REST. Broker outage is invisible to users (notifications delivered when broker recovers). |
| **IV. Performance Requirements** | ✅ PASS | Publish calls use ≤2s timeout context; fail fast, don't block. Consumer processes events sequentially per user (FR-011). No unbounded queries. |
| **Security & Authentication** | ✅ PASS | Broker credentials via `config.yaml` + env vars, never hard-coded. TLS support configurable. No new auth endpoints. |
| **Development Workflow** | ✅ PASS | Commit convention followed. `make check` must pass. Branch name follows `###-short-description` pattern. New adapter packages documented. |

**Gate Result**: ✅ All principles satisfied. No violations to justify.

## Project Structure

### Documentation (this feature)

```text
specs/004-notification-broker-integration/
├── plan.md              # This file
├── research.md          # Phase 0: technology selection & design decisions
├── data-model.md        # Phase 1: entity changes, state transitions, validation
├── quickstart.md        # Phase 1: setup and validation scenarios
├── contracts/           # Phase 1: event schema contract
│   └── notification-event-schema.md
└── tasks.md             # Phase 2: (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/
├── core/
│   ├── domain/
│   │   └── notification.go           # MODIFIED: add EventID field
│   └── ports/
│       └── pubsub.go                 # UNCHANGED: existing Publisher/Subscriber interfaces
├── adapters/
│   └── pubsub/
│       ├── inmemory_pubsub.go        # UNCHANGED: dev fallback
│       ├── inmemory_pubsub_test.go   # UNCHANGED
│       ├── errors.go                 # UNCHANGED
│       ├── kafka_pubsub.go           # NEW: Kafka Publisher + Subscriber
│       ├── kafka_pubsub_test.go      # NEW: Kafka adapter tests
│       ├── rabbitmq_pubsub.go        # NEW: RabbitMQ Publisher + Subscriber
│       ├── rabbitmq_pubsub_test.go   # NEW: RabbitMQ adapter tests
│       └── factory.go                # NEW: broker factory (type → implementation)
├── config/
│   └── config.go                     # MODIFIED: add MessageBrokerConfig
└── api/
    └── handler/
        └── notification_handler.go   # MODIFIED: use event_id from broker metadata

cmd/
└── main.go                           # MODIFIED: use factory instead of direct inmemory construction

configs/
├── config.yaml                       # MODIFIED: add message_broker section
└── config.yaml.example               # MODIFIED: add message_broker example

db/migrations/
├── 000015_add_event_id_to_notifications.up.sql   # NEW
└── 000015_add_event_id_to_notifications.down.sql # NEW

go.mod                                # MODIFIED: add kafka-go + amqp091-go
```

**Structure Decision**: Single backend project. New broker adapters are colocated in `internal/adapters/pubsub/` alongside the existing in-memory implementation. The factory pattern in `factory.go` selects the implementation at startup based on config. A new migration adds the `event_id` column for idempotency.

## Complexity Tracking

> No constitution violations — this section is empty.

## Phase Outputs

### Phase 0: Research — [research.md](./research.md)

- Go Kafka client: `segmentio/kafka-go` (pure Go, no CGO)
- Go RabbitMQ client: `rabbitmq/amqp091-go` (official, stable)
- Both clients cleanly satisfy `ports.Publisher` and `ports.Subscriber`
- Idempotency via DB UNIQUE constraint on `event_id`
- Env-based broker selection with in-memory fallback
- Capped exponential backoff for reconnection
- Schema versioning in event metadata

### Phase 1: Design — [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

- `domain.Notification`: add `EventID string` field
- Migration: partial unique index on `event_id`
- Event schema v1.0: JSON Schema with all required fields + versioning policy
- Config: `MessageBrokerConfig` struct with Kafka/RabbitMQ sub-configs
- Validation scenarios: publish/consume, idempotency, broker outage, graceful shutdown, dead-letter

### Phase 2: Tasks — to be generated by `/speckit.tasks`

