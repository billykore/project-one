# WebSocket Notification Streaming Design

## Goal

Stream real-time notifications to authenticated users via WebSocket, allowing the web app to display live notification updates as they occur. Clients receive only new notifications from connection time forward; historical notifications are retrieved via REST API.

## Requirements

- **Authentication:** JWT Bearer token in HTTP Authorization header during WebSocket upgrade
- **Recipient filtering:** Only send notifications to the intended recipient (by UserID)
- **New notifications only:** Stream from connection time forward; existing notifications accessed via REST API
- **Graceful disconnection:** Handle client disconnects cleanly without goroutine leaks
- **Resilience:** Streamer continues operating even if individual sends fail (broken pipe, connection gone)
- **Integration:** Reuse existing PubSub infrastructure and authentication patterns

## Architecture

### Central Registry Pattern with PubSub Bridge

The system uses three layers:

**1. WebSocket Manager (in-memory registry)**

- Maintains thread-safe mapping: `UserID → *websocket.Conn`
- Provides methods: `Register()`, `Unregister()`, `Send()`, `Close()`
- Prevents goroutine leaks on disconnect

**2. WebSocket Handler (connection lifecycle)**

- HTTP handler at `GET /ws`
- Validates JWT from Authorization header, extracts UserID
- Upgrades HTTP connection to WebSocket
- Registers connection with manager
- Spawns read goroutine to detect disconnects
- On disconnect: unregisters and closes connection

**3. Notification Streamer (pub/sub bridge)**

- Extends existing `NotificationHandler` with `StreamNotifications()` method
- Subscribes to `notifications` topic from PubSub
- For each incoming event:
  - Unmarshals JSON to `domain.Notification`
  - Validates notification
  - Looks up recipient UserID in manager
  - Sends JSON-encoded notification via WebSocket
  - Logs errors (connection gone, send failed) and continues

### Data Flow

```text
1. Client opens WebSocket: GET /ws with Authorization header
2. Handler validates JWT → extracts UserID
3. Handler registers connection: manager.Register(userID, conn)
4. Handler spawns read goroutine to detect disconnects
5. [Async] Notification event published to PubSub
6. Streamer unmarshals event → validates → looks up UserID in manager
7. Streamer sends JSON notification through WebSocket connection
8. Client receives notification in real-time
9. [On disconnect] Read goroutine detects error → unregisters → closes
```

## Components

### File: `internal/adapters/websocket/manager.go`

**Type: Manager**

- Field: `mu sync.RWMutex`
- Field: `connections map[int]*websocket.Conn` (UserID → connection)
- Method: `Register(ctx context.Context, userID int, conn *websocket.Conn) error`
  - Stores connection in map
  - If previous connection exists for userID, closes it gracefully
- Method: `Unregister(userID int)`
  - Removes from map (idempotent)
- Method: `Send(userID int, notification *domain.Notification) error`
  - Looks up connection for userID
  - Returns error if not found (user not connected)
  - Marshals notification to JSON
  - Writes JSON frame via websocket.WriteJSON()
  - Returns error if write fails
- Method: `Close() error`
  - Closes all connections in registry
  - Clears map
- Method: `ConnectionCount() int`
  - Returns current active connections (for monitoring)

### File: `internal/api/handler/websocket_handler.go`

**Type: WebSocketHandler**

- Field: `log ports.Logger`
- Field: `tokenService ports.TokenService`
- Field: `manager *websocket.Manager`
- Field: `validator ports.Validator`
- Constructor: `NewWebSocketHandler(log, tokenService, manager, validator) *WebSocketHandler`
  - Validates all deps are non-nil, panics if missing

- Method: `HandleUpgrade(c echo.Context) error`
  - Extracts Authorization header
  - Validates JWT format (Bearer token)
  - Calls tokenService to validate and extract claims
  - Extracts UserID from claims
  - Returns 401 if auth fails
  - Upgrades HTTP to WebSocket using gorilla/websocket
  - Calls `manager.Register(ctx, userID, conn)`
  - Spawns read goroutine to detect disconnects (reads until error)
  - On any read error, unregisters and closes connection
  - Returns error if upgrade fails
  
- Method: `handleReadLoop(ctx context.Context, userID int, conn *websocket.Conn)`
  - Infinite read loop: `conn.ReadMessage()`
  - On error (client disconnect, closed conn): unregister and return
  - Ignores incoming messages (read loop is for disconnect detection only)

### File: `internal/api/handler/notification_handler.go` (extend existing)

**Add method: `StreamNotifications(ctx context.Context, manager *websocket.Manager) error`**

- Subscribes to `notificationTopic` via `h.subscriber`
- Returns error if subscription fails
- For each received event:
  - Unmarshals event.Payload to `domain.Notification`
  - If unmarshal fails: logs error, continues (bad data in pubsub)
  - Calls `notification.Validate()`
  - If invalid: logs error, continues
  - Calls `manager.Send(notification.UserID, &notification)`
  - If send fails: logs error (connection likely gone), continues
  - If send succeeds: logs info (for debugging)
- Returns only on critical errors (subscription lost)

## Error Handling & Edge Cases

| Scenario | Behavior |
|----------|----------|
| **Invalid JWT on upgrade** | Return HTTP 401, do not upgrade |
| **Malformed token** | Return HTTP 401 |
| **Connection drops mid-read** | Read loop detects error, unregisters gracefully, closes connection |
| **Send to disconnected user** | Streamer logs, continues (user not in registry) |
| **Broken pipe on send** | Streamer logs error, continues |
| **Multiple WS connections for same UserID** | Second registration closes previous connection, replaces it |
| **Invalid notification event** | Streamer logs, skips event |
| **PubSub subscription fails** | StreamNotifications returns error; logged and restarted by caller |
| **Graceful shutdown** | Server closes manager (closes all connections) before exiting |

## Integration Points

1. **PubSub Reuse:** Uses existing `notificationTopic` and subscriber pattern
2. **JWT Validation:** Reuses existing `tokenService.ValidateAndExtractClaims()`
3. **Domain Model:** Works with existing `domain.Notification` struct
4. **Logging:** Logs via existing `ports.Logger` (zerolog)
5. **HTTP Router:** Registers new `GET /ws` route in Echo at startup

## Dependencies & Libraries

- **gorilla/websocket** — WebSocket protocol handling (needs to be added to go.mod)
- **JWT validation** — existing `ports.TokenService`
- **PubSub** — existing subscriber
- **Go standard library** — context, sync, encoding/json

## Testing Strategy

**Unit Tests:**

- Manager: Register/Unregister/Send/Close operations, concurrent access
- WebSocketHandler: JWT validation, upgrade success/failure cases
- NotificationHandler.StreamNotifications: Event unmarshaling, filtering, send failures

**Integration Tests:**

- Full flow: open WS → authenticate → receive notification via pubsub
- Multiple concurrent connections
- Disconnect detection and cleanup
- Error recovery (broken pipes)

## Deployment & Operations

- **New endpoint:** `GET /ws` (no CORS preflights needed for same-origin WS)
- **Environment config:** None required (reuses existing JWT config)
- **Monitoring:** Add metric for active connections via `manager.ConnectionCount()`
- **Logging:** All operations logged at info/error level via zerolog
- **Graceful shutdown:** Manager.Close() called during server shutdown sequence

## Success Criteria

- ✅ Client connects via WebSocket with valid JWT
- ✅ Notification published to PubSub → received by client via WS in <100ms
- ✅ Invalid JWT rejects connection with 401
- ✅ Client disconnect unregisters cleanly (no goroutine leaks)
- ✅ Multiple concurrent connections work independently
- ✅ Notification intended for user A does not go to user B
- ✅ Streamer continues after send failures (resilient)
- ✅ Old notifications remain accessible via REST API
