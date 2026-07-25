# Feature Specification: Notification Broker Integration

**Feature Branch**: `004-notification-broker-integration`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "Improve the notification features in this backend application. The notification expected to be sent to a message broker and consumed by a web app."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reliable Notification Delivery via Message Broker (Priority: P1)

When a user performs an action that generates a notification (someone follows them, likes their post, or comments on their post), the backend publishes the notification event to an external message broker instead of an in-memory queue. The notification consumer — running either within the backend or as a separate process — reads from the broker, persists the notification to the database, and delivers it to the target user via WebSocket in real time. If the consumer is temporarily down, notifications are not lost; they are retained in the broker and processed when the consumer recovers.

**Why this priority**: This is the core architectural change. Switching from an in-memory pubsub to an external message broker ensures notifications survive process restarts and enables decoupled scaling. Without this, the system has no durability guarantees for notification events.

**Independent Test**: Can be fully tested by starting a message broker instance, publishing a notification event, stopping the consumer, publishing another event, restarting the consumer, and verifying both events are processed and persisted. Delivers reliable, durable notification delivery.

**Acceptance Scenarios**:

1. **Given** the backend is running with a connected message broker, **When** a user likes another user's post, **Then** a notification event is published to the broker with the correct type ("like"), actor, and target user information.
2. **Given** the notification consumer is connected to the broker, **When** a notification event arrives on the notification topic/queue, **Then** the consumer persists the notification to the database and pushes it to the target user via WebSocket if they are connected.
3. **Given** the notification consumer is stopped, **When** multiple notification events are published to the broker, **Then** the events are retained in the broker and are not lost.
4. **Given** the notification consumer is restarted after being down, **When** it reconnects to the broker, **Then** all unprocessed notification events are consumed, persisted, and delivered without duplication.

---

### User Story 2 - Decoupled Web App Consumption (Priority: P2)

The web application (or any external consumer) can subscribe directly to notification events from the message broker without going through the backend REST API. This allows the web app to receive notifications within 3 seconds of publication, reducing load on the backend and enabling future architectural flexibility (e.g., push notifications, email digests).

**Why this priority**: Decoupling consumption from the backend is a key architectural benefit of using a message broker. It enables independent scaling of producers and consumers and opens the door to multiple consumer types (web app, email service, push notification service) without modifying the backend.

**Independent Test**: Can be fully tested by running a standalone consumer that subscribes to the broker's notification topic, publishing an event from the backend, and verifying the standalone consumer receives and processes the event correctly. Delivers an independently verifiable consumption path.

**Acceptance Scenarios**:

1. **Given** an external consumer is subscribed to the notification topic on the broker, **When** the backend publishes a notification event, **Then** the external consumer receives the event with all required fields (type, actor, target user, post/comment IDs).
2. **Given** multiple consumers are subscribed to the same notification topic, **When** a single notification event is published, **Then** all consumers receive a copy of the event independently.
3. **Given** the backend is temporarily unavailable, **When** a notification event exists in the broker, **Then** external consumers (once connected and authorized to the broker) can still consume and process the event. Note: direct broker access for external consumers is enabled by the standardized schema and broker configuration but requires infrastructure-level setup (firewall rules, TLS, authentication) that is outside the scope of this application-level feature.

---

### User Story 3 - Notification Event Schema Standardization (Priority: P3)

All notification events published to the message broker follow a consistent, versioned schema. This ensures that any consumer — the backend notification handler, the web app, or future services — can parse and process notification events reliably regardless of which producer generated them.

**Why this priority**: Schema standardization is a prerequisite for reliable multi-consumer architectures. Without it, consumers may break when producers change event formats. It is P3 because the current single-consumer setup works with the existing format; standardization becomes critical when adding multiple consumers.

**Independent Test**: Can be fully tested by validating that all published notification events conform to the defined schema, and that a consumer using only the schema definition can parse events from all three notification types (follow, like, comment). Delivers a contract that decouples producers from consumers.

**Acceptance Scenarios**:

1. **Given** the notification event schema is defined, **When** any notification-producing action occurs (follow, like, comment), **Then** the published event conforms to the schema with all required fields present.
2. **Given** a consumer is built against the schema version N, **When** a producer publishes an event with schema version N, **Then** the consumer parses the event without errors.
3. **Given** the schema is versioned, **When** a new field is added in a backward-compatible manner, **Then** existing consumers continue to function without modification.

---

### Edge Cases

- What happens when the message broker is unreachable during backend startup? The system should retry the connection with backoff and log warnings, but not block API request handling.
- What happens when the broker connection is lost mid-operation? Publish operations should fail gracefully with timeouts; the use case should log the error and continue (notification delivery is best-effort, not transactional).
- How does the system handle duplicate events (at-least-once delivery)? The notification consumer must detect and skip duplicate notifications (e.g., by idempotency key or unique constraint).
- What happens when a notification event contains malformed or invalid data? The consumer should move malformed messages to a dead-letter queue/topic for inspection rather than retrying indefinitely.
- How does the system handle broker authentication and authorization? The broker connection must use credentials; these must be configurable and never hard-coded.
- What happens to in-flight notifications during a graceful shutdown? The publisher must flush pending messages before closing; the consumer must finish processing the current event before disconnecting.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST publish notification events (follow, like, comment) to an external message broker instead of the current in-memory pubsub.
- **FR-002**: System MUST retain the `ports.Publisher` and `ports.Subscriber` interfaces as the abstraction layer, with the new broker implementation satisfying these interfaces.
- **FR-003**: The message broker implementation MUST support at-least-once delivery semantics with message persistence (messages survive broker restarts).
- **FR-004**: The notification consumer MUST be idempotent — processing the same event multiple times MUST NOT result in duplicate notifications being persisted or delivered.
- **FR-005**: System MUST support configurable broker connection parameters (host, port, credentials, virtual host/topic prefix) via the existing `config.yaml` configuration file.
- **FR-006**: System MUST log all broker connection events (connect, disconnect, reconnect) at appropriate log levels (Info for normal operations, Warn for transient failures, Error for persistent failures).
- **FR-007**: Notification events published to the broker MUST include a unique event identifier to support idempotent consumption and tracing.
- **FR-008**: Notification events published to the broker MUST include all required fields: event type (follow/like/comment), actor ID, actor username, target user ID, and relevant entity IDs (post ID for likes/comments, comment ID for comments).
- **FR-009**: System MUST handle broker unavailability gracefully — publish failures MUST NOT cause API requests to fail; they MUST be logged and the user-facing operation MUST still succeed.
- **FR-010**: System MUST provide a health check or status endpoint that indicates whether the broker connection is healthy.
- **FR-011**: The broker consumer MUST process events sequentially per target user to preserve notification ordering (notifications for user A are ordered, but may be processed concurrently with notifications for user B).
- **FR-012**: System MUST support graceful shutdown — the publisher MUST flush pending messages and the consumer MUST complete in-flight processing before closing the broker connection.
- **FR-013**: Malformed or unprocessable events MUST be routed to a dead-letter mechanism (dead-letter queue or equivalent) after a configurable number of retry attempts (default: 3), rather than being retried indefinitely.

### Key Entities

- **Notification Event**: Represents a notification published to the message broker. Contains event ID, event type (follow/like/comment), actor information, target user information, related entity references (post, comment), and schema version.
- **Notification (Domain Entity)**: The persisted notification record, as currently defined in `domain.Notification`. Stored in the database after the consumer processes a broker event.
- **Broker Connection Configuration**: Connection parameters for the message broker, including host, port, authentication credentials, virtual host or topic prefix, reconnection settings, and TLS options.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Notification events survive backend process restarts — 100% of events published before a restart are processed after the system recovers (zero message loss under normal operating conditions).
- **SC-002**: Notification delivery latency from user action to WebSocket push remains under 2 seconds at p95 under normal load (comparable to or better than the current in-memory implementation).
- **SC-003**: The system handles at least 1,000 notification events per minute without message loss or degradation in API response times.
- **SC-004**: Duplicate notification events are detected and skipped with 100% accuracy — no user ever sees the same notification twice due to at-least-once redelivery.
- **SC-005**: Broker connection failures do not cause API errors — 100% of user-facing operations (follow, like, comment) succeed even when the broker is temporarily unavailable.
- **SC-006**: A new consumer can be added to consume notification events from the broker without any changes to the producer (backend) code or configuration.

## Assumptions

- The message broker technology is chosen based on operational requirements. The spec is technology-agnostic and applies to any message broker that supports publish-subscribe or queue-based messaging with message persistence (e.g., RabbitMQ, NATS, Kafka, Redis Streams).
- The existing in-memory pubsub implementation serves as a development/testing fallback. The broker implementation is selected and configured per environment (in-memory for local dev, external broker for staging/production).
- The web app's consumption of broker events is assumed to be via the backend's existing WebSocket or REST API for the initial implementation. Direct broker consumption by the web app (e.g., via WebSocket or SSE from the broker) is a future architectural option enabled by this work.
- The existing notification types (follow, like, comment) are the only event types in scope. New notification types will follow the same schema and publishing pattern.
- The broker is deployed and managed as infrastructure; this feature covers the application-level integration (publishing, consuming, configuration) but not broker cluster setup or operation.
- Connection security (TLS) for the broker is configured at the infrastructure level; the application supports TLS configuration via `config.yaml` but does not manage certificates.
