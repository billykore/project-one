package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	sseadapter "github.com/billykore/project-one/internal/adapters/sse"
	wsadapter "github.com/billykore/project-one/internal/adapters/websocket"
	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNotificationHandler_StreamNotifications(t *testing.T) {
	t.Parallel()

	sseManager := sseadapter.NewManager()
	h := NewNotificationHandler(
		noopLogger{},
		nil,
		nil,
		staticUserUseCase{user: &domain.User{ID: 42, Username: "alice"}},
		nil,
		nil,
		sseManager,
	)

	e := echo.New()
	reqCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest("GET", "/notifications/stream", nil).WithContext(reqCtx)
	rec := httptest.NewRecorder()
	// Wrap with a mutex-protected writer to prevent data races between
	// the handler goroutine writing to rec and the Eventually callback
	// reading rec.Body concurrently.
	locked := &lockedWriter{w: rec}
	c := e.NewContext(req, locked)
	c.Set("username", "alice")

	errCh := make(chan error, 1)
	go func() {
		errCh <- h.StreamNotifications(c)
	}()

	require.Eventually(t, func() bool {
		return sseManager.ConnectionCount(42) == 1
	}, time.Second, 10*time.Millisecond)

	require.NoError(t, sseManager.Send(&dto.NotificationResponse{
		ID:            9,
		UserID:        42,
		ActorID:       7,
		ActorUsername: "bob",
		Type:          "follow",
		IsRead:        false,
		Title:         "New Follower",
		Body:          "bob started following you.",
	}))

	var payload dto.NotificationResponse

	require.Eventually(t, func() bool {
		body := locked.bodyString()
		idx := strings.Index(body, "data: ")
		if idx < 0 {
			return false
		}
		line := body[idx:]
		line = strings.SplitN(line, "\n", 2)[0]
		if !strings.HasPrefix(line, "data: ") {
			return false
		}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(strings.TrimSpace(line), "data: ")), &payload); err != nil {
			return false
		}
		return true
	}, time.Second, 10*time.Millisecond)

	cancel()
	require.NoError(t, <-errCh)

	assert.Equal(t, 9, payload.ID)
	assert.Equal(t, 42, payload.UserID)
	assert.Equal(t, "bob", payload.ActorUsername)
	assert.Equal(t, "follow", payload.Type)
}

func TestNotificationHandler_ListenSkipsMissingSSEClient(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriber := mocks.NewMockSubscriber(ctrl)
	notificationUc := mocks.NewMockNotificationUseCase(ctrl)
	wsManager := wsadapter.NewManager()
	sseManager := sseadapter.NewManager()

	h := NewNotificationHandler(
		noopLogger{},
		subscriber,
		notificationUc,
		nil,
		nil,
		wsManager,
		sseManager,
	)

	notificationEvent := domain.NotificationEvent{
		EventID:       "backend-1",
		SchemaVersion: "1.0",
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Notification: domain.Notification{
			UserID:        42,
			ActorID:       7,
			ActorUsername: "bob",
			Type:          domain.NotificationTypeFollow,
			CreatedAt:     time.Now().UTC(),
		},
	}
	payload, err := json.Marshal(notificationEvent)
	require.NoError(t, err)

	notificationUc.EXPECT().
		SaveNotification(gomock.Any(), gomock.Any()).
		DoAndReturn(func(context.Context, *domain.Notification) error {
			return nil
		})

	subscriber.EXPECT().
		Subscribe(gomock.Any(), "notifications", gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ string, handler ports.EventHandler) error {
			return handler(ctx, ports.Event{Payload: payload})
		})

	require.NoError(t, h.Listen(context.Background()))
}

type noopLogger struct{}

func (noopLogger) Debug(context.Context, string, ...any) {}
func (noopLogger) Info(context.Context, string, ...any)  {}
func (noopLogger) Warn(context.Context, string, ...any)  {}
func (noopLogger) Error(context.Context, string, ...any) {}
func (noopLogger) Fatal(context.Context, string, ...any) {}

type staticUserUseCase struct {
	user *domain.User
}

func (u staticUserUseCase) Register(context.Context, *domain.User) error {
	return nil
}

func (u staticUserUseCase) GetUser(context.Context, string) (*domain.User, error) {
	return u.user, nil
}

func (u staticUserUseCase) ChangePassword(context.Context, string, string, string) error {
	return nil
}

func (u staticUserUseCase) UpdateProfile(context.Context, string, *domain.User) error {
	return nil
}

var _ ports.UserUseCase = staticUserUseCase{}

// lockedWriter wraps an http.ResponseWriter with a mutex to allow safe
// concurrent writes (from a handler goroutine) and reads (from a test
// polling body content).
type lockedWriter struct {
	mu sync.Mutex
	w  http.ResponseWriter
}

func (l *lockedWriter) Header() http.Header        { return l.w.Header() }
func (l *lockedWriter) WriteHeader(statusCode int) { l.w.WriteHeader(statusCode) }

func (l *lockedWriter) Write(b []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(b)
}

// Flush delegates to the underlying writer. httptest.ResponseRecorder.Flush
// is a no-op so no lock is required.
func (l *lockedWriter) Flush() {
	l.w.(http.Flusher).Flush()
}

func (l *lockedWriter) bodyString() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.(*httptest.ResponseRecorder).Body.String()
}
