# Notification Event Schema (v1.0)

**Contract Type**: Message Schema  
**Transport**: Message Broker (Kafka / RabbitMQ)  
**Format**: JSON  
**Serialization**: `application/json` with UTF-8 encoding  

## Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://api.project1.com/schemas/notification-event-v1.json",
  "title": "Notification Event",
  "description": "A notification event published to the message broker when a social action occurs (follow, like, comment).",
  "type": "object",
  "required": [
    "event_id",
    "schema_version",
    "type",
    "user_id",
    "actor_id",
    "actor_username",
    "created_at",
    "timestamp"
  ],
  "properties": {
    "event_id": {
      "type": "string",
      "description": "Unique event identifier. Format: {producer}-{uuid}. Used for consumer idempotency.",
      "examples": ["backend-a1b2c3d4e5f6"]
    },
    "schema_version": {
      "type": "string",
      "description": "Schema version for this event. Consumers use this to select the correct deserializer.",
      "examples": ["1.0"]
    },
    "type": {
      "type": "string",
      "enum": ["follow", "like", "comment"],
      "description": "Notification event type."
    },
    "user_id": {
      "type": "integer",
      "minimum": 1,
      "description": "Target user ID — the notification recipient."
    },
    "actor_id": {
      "type": "integer",
      "minimum": 1,
      "description": "Actor user ID — the user who triggered the notification."
    },
    "actor_username": {
      "type": "string",
      "minLength": 1,
      "description": "Username of the actor."
    },
    "post_id": {
      "type": "integer",
      "minimum": 0,
      "description": "Related post ID. 0 if not applicable (follow notifications)."
    },
    "comment_id": {
      "type": "integer",
      "minimum": 0,
      "description": "Related comment ID. 0 if not applicable (follow, like notifications)."
    },
    "created_at": {
      "type": "string",
      "format": "date-time",
      "description": "ISO 8601 timestamp of when the notification-triggering action occurred."
    },
    "timestamp": {
      "type": "string",
      "format": "date-time",
      "description": "ISO 8601 timestamp of when the event was published to the broker."
    }
  },
  "additionalProperties": false
}
```

## Examples

### Follow Notification

```json
{
  "event_id": "backend-f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "schema_version": "1.0",
  "type": "follow",
  "user_id": 42,
  "actor_id": 7,
  "actor_username": "alice",
  "post_id": 0,
  "comment_id": 0,
  "created_at": "2026-07-25T10:30:00Z",
  "timestamp": "2026-07-25T10:30:00.123Z"
}
```

### Like Notification

```json
{
  "event_id": "backend-d94b81a2-8412-4f1e-9c3b-9e46a6b7c8d9",
  "schema_version": "1.0",
  "type": "like",
  "user_id": 42,
  "actor_id": 15,
  "actor_username": "bob",
  "post_id": 108,
  "comment_id": 0,
  "created_at": "2026-07-25T11:15:00Z",
  "timestamp": "2026-07-25T11:15:00.456Z"
}
```

### Comment Notification

```json
{
  "event_id": "backend-3e8f1a6b-d254-4a7c-b91a-f38c2d5e6f70",
  "schema_version": "1.0",
  "type": "comment",
  "user_id": 42,
  "actor_id": 23,
  "actor_username": "charlie",
  "post_id": 108,
  "comment_id": 512,
  "created_at": "2026-07-25T11:20:00Z",
  "timestamp": "2026-07-25T11:20:00.789Z"
}
```

## Event Metadata (Broker Headers)

In addition to the JSON payload, events carry the following broker headers/metadata:

| Header Key | Type | Description |
|-----------|------|-------------|
| `event_id` | `string` | Same as `event_id` in payload — duplicated for broker-level routing/filtering |
| `schema_version` | `string` | Schema version for header-based routing |
| `event_type` | `string` | `follow` / `like` / `comment` — enables topic-partitioning by type |
| `producer` | `string` | Identifies the producing service (always `"project1-backend"`) |

## Versioning Policy

- **MAJOR** (e.g., 1.0 → 2.0): Breaking changes — removed or renamed required fields, changed field types. New consumer required.
- **MINOR** (e.g., 1.0 → 1.1): New optional fields added. Backward-compatible; old consumers ignore new fields.
- **PATCH** (e.g., 1.0 → 1.0.1): Documentation or validation constraint clarifications. No payload changes.

Consumers MUST check `schema_version` and handle unknown versions gracefully (log a warning and skip, or route to dead-letter).
