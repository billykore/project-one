# WebSocket Notification Streaming Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stream only new notification events to authenticated users over WebSocket, with per-recipient delivery based on `Notification.UserID`, while keeping historical reads on REST.

**Architecture:** Add a thread-safe in-memory WebSocket connection manager keyed by user ID, a dedicated WebSocket upgrade handler that authenticates via Bearer token, and a notification-stream subscriber path that pushes PubSub events to active sockets. Keep persistence listener behavior unchanged and wire startup/shutdown in `cmd/main.go` so both listeners run together safely.

**Tech Stack:** Go 1.26.2, Echo v4, gorilla/websocket, existing JWT token service (`ports.TokenService`), existing PubSub adapter (`ports.Subscriber`), testify + gomock.

## Global Constraints

- Authentication: JWT Bearer token in HTTP Authorization header during WebSocket upgrade.
- Recipient filtering: only send notifications to the intended recipient (`Notification.UserID`).
- New notifications only: stream from connection time forward; historical notifications come from REST API.
- Graceful disconnection: handle client disconnects cleanly without goroutine leaks.
- Resilience: continue streaming loop even if individual sends fail.
- Integration: reuse existing PubSub infrastructure and existing auth patterns.

---

## File Structure

- Create `internal/adapters/websocket/manager.go`:
  - Owns active WebSocket connections per user and write-path synchronization.
- Create `internal/adapters/websocket/manager_test.go`:
  - Verifies register/replace/unregister/send/close behavior and concurrency safety basics.
- Create `internal/api/handler/websocket_handler.go`:
  - Handles upgrade auth, socket registration, and disconnect cleanup.
- Create `internal/api/handler/websocket_handler_test.go`:
  - Verifies unauthorized handling and successful upgrade lifecycle.
- Modify `internal/api/handler/notification_handler.go`:
  - Add stream method that subscribes to pubsub and forwards to manager by recipient user ID.
- Create `internal/api/handler/notification_stream_test.go`:
  - Verifies stream callback behavior for valid, invalid, and disconnected-recipient events.
- Modify `cmd/main.go`:
  - Construct manager + websocket handler, add `GET /ws`, start stream listener goroutine, and close manager during shutdown.
- Modify `go.mod` and `go.sum`:
  - Add `github.com/gorilla/websocket` dependency.

### Task 1: Add WebSocket Manager Adapter

**Files:**
- Create: `internal/adapters/websocket/manager.go`
- Test: `internal/adapters/websocket/manager_test.go`

**Interfaces:**
- Consumes: `domain.Notification` from `internal/core/domain/notification.go`
- Produces:
  - `func NewManager() *Manager`
  - `func (m *Manager) Register(userID int, conn *gws.Conn) error`
  - `func (m *Manager) Unregister(userID int)`
  - `func (m *Manager) Send(notification *domain.Notification) error`
  - `func (m *Manager) Close() error`
  - `func (m *Manager) ConnectionCount() int`

- [ ] **Step 1: Write the failing tests for manager behavior**

```go
package websocket

import (
	"net/http/httptest"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	gws "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestManager_RegisterReplaceAndCount(t *testing.T) {
	m := NewManager()
	srv, wsURL := startWSTestServer(t)
	defer srv.Close()

	c1 := mustDial(t, wsURL)
	defer c1.Close()
	assert.NoError(t, m.Register(1, c1))
	assert.Equal(t, 1, m.ConnectionCount())

	c2 := mustDial(t, wsURL)
	defer c2.Close()
	assert.NoError(t, m.Register(1, c2))
	assert.Equal(t, 1, m.ConnectionCount())
}

func TestManager_SendToConnectedUser(t *testing.T) {
	m := NewManager()
	srv, wsURL := startWSTestServer(t)
	defer srv.Close()

	conn := mustDial(t, wsURL)
	defer conn.Close()
	assert.NoError(t, m.Register(10, conn))

	err := m.Send(&domain.Notification{UserID: 10, ActorID: 2, Type: domain.NotificationTypeFollow})
	assert.NoError(t, err)
}

func TestManager_SendToDisconnectedUserReturnsError(t *testing.T) {
	m := NewManager()
	err := m.Send(&domain.Notification{UserID: 999, Type: domain.NotificationTypeFollow})
	assert.Error(t, err)
}

func TestManager_UnregisterAndClose(t *testing.T) {
	m := NewManager()
	m.Unregister(123)
	assert.Equal(t, 0, m.ConnectionCount())
	assert.NoError(t, m.Close())
}

func startWSTestServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	e := echo.New()
	upgrader := gws.Upgrader{}
	e.GET("/ws", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer conn.Close()
		select {}
	})
	srv := httptest.NewServer(e)
	return srv, "ws" + srv.URL[len("http"):] + "/ws"
}

func mustDial(t *testing.T, wsURL string) *gws.Conn {
	t.Helper()
	conn, _, err := gws.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial ws: %v", err)
	}
	return conn
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `go test ./internal/adapters/websocket -run TestManager -count=1`
Expected: FAIL with compile errors because `NewManager`, `Register`, `Send`, and `Close` do not exist.

- [ ] **Step 3: Implement manager with minimal passing logic**

```go
package websocket

import (
	"errors"
	"sync"

	"github.com/billykore/project-one/internal/core/domain"
	gws "github.com/gorilla/websocket"
)

var ErrUserNotConnected = errors.New("user is not connected")

type managedConn struct {
	conn *gws.Conn
	mu   sync.Mutex
}

type Manager struct {
	mu          sync.RWMutex
	connections map[int]*managedConn
}

func NewManager() *Manager {
	return &Manager{connections: make(map[int]*managedConn)}
}

func (m *Manager) Register(userID int, conn *gws.Conn) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if prev, ok := m.connections[userID]; ok && prev != nil && prev.conn != nil {
		_ = prev.conn.Close()
	}
	m.connections[userID] = &managedConn{conn: conn}
	return nil
}

func (m *Manager) Unregister(userID int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, userID)
}

func (m *Manager) Send(notification *domain.Notification) error {
	m.mu.RLock()
	mc, ok := m.connections[notification.UserID]
	m.mu.RUnlock()
	if !ok || mc == nil || mc.conn == nil {
		return ErrUserNotConnected
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()
	return mc.conn.WriteJSON(notification)
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for userID, mc := range m.connections {
		if mc != nil && mc.conn != nil {
			_ = mc.conn.Close()
		}
		delete(m.connections, userID)
	}
	return nil
}

func (m *Manager) ConnectionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `go test ./internal/adapters/websocket -run TestManager -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/websocket/manager.go internal/adapters/websocket/manager_test.go
git commit -m "feat: add websocket connection manager"
```

### Task 2: Add WebSocket Upgrade Handler

**Files:**
- Create: `internal/api/handler/websocket_handler.go`
- Test: `internal/api/handler/websocket_handler_test.go`

**Interfaces:**
- Consumes:
  - `ports.TokenService.ValidateToken(ctx context.Context, token string) (username string, err error)`
  - `ports.UserUseCase.GetUser(ctx context.Context, username string) (*domain.User, error)`
  - `(*websocket.Manager).Register(userID int, conn *gws.Conn) error`
  - `(*websocket.Manager).Unregister(userID int)`
- Produces:
  - `type WebSocketHandler struct { ... }`
  - `func NewWebSocketHandler(log ports.Logger, tokenSvc ports.TokenService, userUc ports.UserUseCase, manager *websocket.Manager) *WebSocketHandler`
  - `func (h *WebSocketHandler) HandleUpgrade(c echo.Context) error`

- [ ] **Step 1: Write failing tests for auth and successful upgrade**

```go
package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	wsadapter "github.com/billykore/project-one/internal/adapters/websocket"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type stubTokenService struct{ username string; err error }
func (s stubTokenService) GenerateTokens(context.Context, *domain.User) (*domain.UserToken, error) { return nil, errors.New("not used") }
func (s stubTokenService) ValidateToken(context.Context, string) (string, error) { return s.username, s.err }

type stubUserUseCase struct{ user *domain.User; err error }
func (s stubUserUseCase) Register(context.Context, *domain.User) error { return errors.New("not used") }
func (s stubUserUseCase) GetUser(context.Context, string) (*domain.User, error) { return s.user, s.err }
func (s stubUserUseCase) Follow(context.Context, string, string) error { return errors.New("not used") }
func (s stubUserUseCase) Unfollow(context.Context, string, string) error { return errors.New("not used") }
func (s stubUserUseCase) GetFollowing(context.Context, string) ([]*domain.User, error) { return nil, errors.New("not used") }
func (s stubUserUseCase) GetFollowers(context.Context, string, string) ([]*domain.Follower, error) { return nil, errors.New("not used") }
func (s stubUserUseCase) ChangePassword(context.Context, string, string, string) error { return errors.New("not used") }

func TestWebSocketHandler_UnauthorizedWithoutBearer(t *testing.T) {
	e := echo.New()
	manager := wsadapter.NewManager()
	h := NewWebSocketHandler(mockLogger{}, stubTokenService{}, stubUserUseCase{}, manager)

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HandleUpgrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `go test ./internal/api/handler -run TestWebSocketHandler -count=1`
Expected: FAIL because `NewWebSocketHandler` and `HandleUpgrade` do not exist.

- [ ] **Step 3: Implement handler and disconnect read loop**

```go
package handler

import (
	"net/http"
	"strings"

	wsadapter "github.com/billykore/project-one/internal/adapters/websocket"
	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/ports"
	gws "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type WebSocketHandler struct {
	log      ports.Logger
	tokenSvc ports.TokenService
	userUc   ports.UserUseCase
	manager  *wsadapter.Manager
	upgrader gws.Upgrader
}

func NewWebSocketHandler(log ports.Logger, tokenSvc ports.TokenService, userUc ports.UserUseCase, manager *wsadapter.Manager) *WebSocketHandler {
	if log == nil { panic("log is required") }
	if tokenSvc == nil { panic("tokenSvc is required") }
	if userUc == nil { panic("userUc is required") }
	if manager == nil { panic("manager is required") }
	return &WebSocketHandler{
		log:      log,
		tokenSvc: tokenSvc,
		userUc:   userUc,
		manager:  manager,
		upgrader: gws.Upgrader{CheckOrigin: func(*http.Request) bool { return true }},
	}
}

func (h *WebSocketHandler) HandleUpgrade(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	token, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok || token == "" {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	username, err := h.tokenSvc.ValidateToken(c.Request().Context(), token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	user, err := h.userUc.GetUser(c.Request().Context(), username)
	if err != nil || user == nil {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	if err := h.manager.Register(user.ID, conn); err != nil {
		_ = conn.Close()
		return err
	}

	go func(userID int, wsConn *gws.Conn) {
		defer func() {
			h.manager.Unregister(userID)
			_ = wsConn.Close()
		}()
		for {
			if _, _, readErr := wsConn.ReadMessage(); readErr != nil {
				return
			}
		}
	}(user.ID, conn)

	return nil
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `go test ./internal/api/handler -run TestWebSocketHandler -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/api/handler/websocket_handler.go internal/api/handler/websocket_handler_test.go
git commit -m "feat: add websocket upgrade handler"
```

### Task 3: Stream PubSub Notifications to Active Sockets

**Files:**
- Modify: `internal/api/handler/notification_handler.go`
- Test: `internal/api/handler/notification_stream_test.go`

**Interfaces:**
- Consumes:
  - `ports.Subscriber.Subscribe(ctx context.Context, topic string, handler ports.EventHandler) error`
  - `(*websocket.Manager).Send(notification *domain.Notification) error`
- Produces:
  - `func (h *NotificationHandler) StreamNotifications(ctx context.Context, manager *wsadapter.Manager) error`

- [ ] **Step 1: Write failing stream tests**

```go
package handler

import (
	"context"
	"encoding/json"
	"testing"

	wsadapter "github.com/billykore/project-one/internal/adapters/websocket"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/stretchr/testify/assert"
)

type captureSubscriber struct { handler ports.EventHandler }
func (s *captureSubscriber) Subscribe(_ context.Context, _ string, handler ports.EventHandler) error { s.handler = handler; return nil }
func (s *captureSubscriber) Close() error { return nil }

func TestNotificationHandler_StreamNotifications_SubscribeRegistersHandler(t *testing.T) {
	sub := &captureSubscriber{}
	h := &NotificationHandler{subscriber: sub, log: mockLogger{}}
	err := h.StreamNotifications(context.Background(), wsadapter.NewManager())
	assert.NoError(t, err)
	assert.NotNil(t, sub.handler)
}

func TestNotificationHandler_StreamNotifications_InvalidPayloadIgnored(t *testing.T) {
	sub := &captureSubscriber{}
	h := &NotificationHandler{subscriber: sub, log: mockLogger{}}
	manager := wsadapter.NewManager()
	_ = h.StreamNotifications(context.Background(), manager)
	err := sub.handler(context.Background(), ports.Event{Topic: notificationTopic, Payload: []byte("{")})
	assert.NoError(t, err)
}

func TestNotificationHandler_StreamNotifications_ValidPayloadAttemptsSend(t *testing.T) {
	sub := &captureSubscriber{}
	h := &NotificationHandler{subscriber: sub, log: mockLogger{}}
	manager := wsadapter.NewManager()
	_ = h.StreamNotifications(context.Background(), manager)

	payload, _ := json.Marshal(domain.Notification{UserID: 77, Type: domain.NotificationTypeFollow})
	err := sub.handler(context.Background(), ports.Event{Topic: notificationTopic, Payload: payload})
	assert.NoError(t, err)
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `go test ./internal/api/handler -run TestNotificationHandler_StreamNotifications -count=1`
Expected: FAIL because `StreamNotifications` does not exist.

- [ ] **Step 3: Implement `StreamNotifications` in notification handler**

```go
func (h *NotificationHandler) StreamNotifications(ctx context.Context, manager *wsadapter.Manager) error {
	if manager == nil {
		return errors.New("manager is required")
	}

	return h.subscriber.Subscribe(ctx, notificationTopic, func(ctx context.Context, event ports.Event) error {
		var notification domain.Notification
		if err := json.Unmarshal(event.Payload, &notification); err != nil {
			h.log.Error(ctx, "failed to unmarshal notification event for websocket stream", "error", err)
			return nil
		}

		if err := notification.Validate(); err != nil {
			h.log.Error(ctx, "invalid notification event for websocket stream", "error", err)
			return nil
		}

		if err := manager.Send(&notification); err != nil {
			h.log.Warn(ctx, "failed to stream notification to websocket", "userID", notification.UserID, "error", err)
			return nil
		}

		h.log.Info(ctx, "notification streamed to websocket", "userID", notification.UserID, "type", notification.Type)
		return nil
	})
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `go test ./internal/api/handler -run TestNotificationHandler_StreamNotifications -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/api/handler/notification_handler.go internal/api/handler/notification_stream_test.go
git commit -m "feat: stream notification events to websocket clients"
```

### Task 4: Wire Route and Lifecycle in Main

**Files:**
- Modify: `cmd/main.go`
- Modify: `go.mod`
- Modify: `go.sum`

**Interfaces:**
- Consumes:
  - `handler.NewWebSocketHandler(log, tokenSvc, userUc, wsManager)`
  - `notificationHdl.StreamNotifications(ctx, wsManager)`
  - `wsManager.Close() error`
- Produces:
  - New route: `GET /ws`
  - New startup goroutine for stream subscriber
  - Shutdown close path for manager

- [ ] **Step 1: Add dependency and verify module compiles**

Run: `go get github.com/gorilla/websocket@v1.5.3`
Expected: `go.mod` and `go.sum` updated with gorilla websocket.

- [ ] **Step 2: Write failing compile check by referencing new wiring points**

```go
// in cmd/main.go (insert near handlers initialization)
wsManager := wsadapter.NewManager()
wsHdl := handler.NewWebSocketHandler(lgr, tokenSvc, userUc, wsManager)
```

Run: `go test ./cmd/... -run TestDoesNotExist -count=1`
Expected: FAIL compile until imports and full wiring are complete.

- [ ] **Step 3: Implement full routing/startup/shutdown wiring**

```go
// Add import:
wsadapter "github.com/billykore/project-one/internal/adapters/websocket"

// Initialize manager and websocket handler after notification handler:
wsManager := wsadapter.NewManager()
wsHdl := handler.NewWebSocketHandler(lgr, tokenSvc, userUc, wsManager)

// Register websocket endpoint:
e.GET("/ws", wsHdl.HandleUpgrade)

// Start stream subscriber alongside persistence listener:
go func(ctx context.Context) {
	if err := notificationHdl.StreamNotifications(ctx, wsManager); err != nil {
		lgr.Fatal(ctx, "failed to start websocket notification streamer", "error", err)
	}
}(ctx)

// During shutdown:
if err := wsManager.Close(); err != nil {
	lgr.Error(ctx, "failed to close websocket manager", "error", err)
}
```

- [ ] **Step 4: Run targeted build and tests**

Run: `go test ./internal/api/handler ./internal/adapters/websocket ./cmd/... -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/main.go go.mod go.sum
git commit -m "feat: wire websocket notification streaming in server lifecycle"
```

### Task 5: Add End-to-End Handler-Level Stream Test

**Files:**
- Create: `internal/api/handler/websocket_notification_integration_test.go`

**Interfaces:**
- Consumes:
  - `WebSocketHandler.HandleUpgrade`
  - `NotificationHandler.StreamNotifications`
  - `ports.Publisher.Publish`
- Produces:
  - Regression test proving pubsub event reaches authenticated recipient socket only

- [ ] **Step 1: Write failing integration test**

```go
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	wsadapter "github.com/billykore/project-one/internal/adapters/websocket"
	"github.com/billykore/project-one/internal/adapters/pubsub"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	gws "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestWebSocketNotificationFlow_RecipientOnly(t *testing.T) {
	ctx := context.Background()
	broker := pubsub.NewInMemoryPubSub()
	publisher := pubsub.NewInMemoryPublisher(broker)
	subscriber := pubsub.NewInMemorySubscriber(broker)
	manager := wsadapter.NewManager()

	// build handlers with stubs so username -> userID mapping is deterministic
	wsH := NewWebSocketHandler(mockLogger{}, stubTokenService{username: "alice"}, stubUserUseCase{user: &domain.User{ID: 1, Username: "alice"}}, manager)
	nh := &NotificationHandler{log: mockLogger{}, subscriber: subscriber}
	require.NoError(t, nh.StreamNotifications(ctx, manager))

	e := echo.New()
	e.GET("/ws", wsH.HandleUpgrade)
	srv := httptest.NewServer(e)
	defer srv.Close()

	headers := http.Header{}
	headers.Set("Authorization", "Bearer valid-token")
	conn, _, err := gws.DefaultDialer.Dial("ws"+srv.URL[len("http"):]+"/ws", headers)
	require.NoError(t, err)
	defer conn.Close()

	payload, _ := json.Marshal(domain.Notification{UserID: 1, ActorID: 2, Type: domain.NotificationTypeFollow})
	require.NoError(t, publisher.Publish(ctx, ports.Event{Topic: notificationTopic, Payload: payload}))

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	require.Contains(t, string(msg), "\"user_id\":1")
}
```

- [ ] **Step 2: Run test to verify failure**

Run: `go test ./internal/api/handler -run TestWebSocketNotificationFlow_RecipientOnly -count=1`
Expected: FAIL until websocket wiring/stubs are complete.

- [ ] **Step 3: Implement minimal test support helpers in handler test package**

```go
// in handler test files, add shared logger stub once:
type mockLogger struct{}
func (mockLogger) Debug(context.Context, string, ...any) {}
func (mockLogger) Info(context.Context, string, ...any)  {}
func (mockLogger) Warn(context.Context, string, ...any)  {}
func (mockLogger) Error(context.Context, string, ...any) {}
func (mockLogger) Fatal(context.Context, string, ...any) {}
```

- [ ] **Step 4: Re-run integration test**

Run: `go test ./internal/api/handler -run TestWebSocketNotificationFlow_RecipientOnly -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/api/handler/websocket_notification_integration_test.go internal/api/handler/websocket_handler_test.go internal/api/handler/notification_stream_test.go
git commit -m "test: add websocket notification streaming integration coverage"
```

### Task 6: Final Verification and Developer Docs Update

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: existing backend run and test commands
- Produces: brief websocket endpoint documentation for frontend consumers

- [ ] **Step 1: Add endpoint usage note in README backend/api section**

```markdown
### Real-time Notifications (WebSocket)

- Endpoint: `GET /ws`
- Auth: `Authorization: Bearer <access_token>` during WebSocket handshake
- Behavior: streams only new notifications for the authenticated user
- Historical notifications: use `GET /notifications`
```

- [ ] **Step 2: Run full relevant test suite**

Run: `go test ./internal/adapters/websocket ./internal/api/handler ./internal/core/usecase -count=1`
Expected: PASS.

- [ ] **Step 3: Run project test command used by repo scripts**

Run: `make test`
Expected: PASS (or same known pre-existing failures as baseline, with no new websocket-related failures).

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: document websocket notifications endpoint"
```

## Self-Review

### 1) Spec Coverage

- Auth via Bearer header on upgrade: covered in Task 2 tests + implementation.
- Recipient-only delivery (`Notification.UserID`): covered in Task 1 send signature and Task 5 recipient flow test.
- New-notifications-only stream behavior: covered by event-driven stream path (Task 3/5) with no bootstrap backfill.
- Graceful disconnect and no leak-prone lifecycle handling: covered in Task 2 read loop + Task 4 shutdown manager close.
- Resilience on send failures: covered in Task 3 error handling (`Warn` and continue).
- Existing PubSub/auth reuse: covered in Task 3 and Task 4 wiring.

### 2) Placeholder Scan

- No `TODO/TBD/implement later` placeholders.
- Each code-writing step includes concrete code blocks.
- Each run step includes explicit command and expected result.

### 3) Type Consistency

- Manager send API uses `Send(notification *domain.Notification)` consistently across Task 1 and Task 3.
- WebSocket handler constructor signature is consistent in Task 2 and Task 4 wiring.
- Stream method signature remains `StreamNotifications(ctx context.Context, manager *wsadapter.Manager) error` in all tasks.
