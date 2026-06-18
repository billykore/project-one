package websocket

import (
	"net/http/httptest"
	"testing"

	"github.com/billykore/project-one/internal/api/dto"
	gws "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestManager_RegisterReplaceAndCount(t *testing.T) {
	m := NewManager()
	srv, wsURL := startWSTestServer(t)
	defer srv.Close()

	c1 := mustDial(t, wsURL)
	defer func() {
		assert.NoError(t, c1.Close())
	}()
	assert.NoError(t, m.Register(1, c1))
	assert.Equal(t, 1, m.ConnectionCount())

	c2 := mustDial(t, wsURL)
	defer func() {
		assert.NoError(t, c2.Close())
	}()
	assert.NoError(t, m.Register(1, c2))
	assert.Equal(t, 1, m.ConnectionCount())
}

func TestManager_SendToConnectedUser(t *testing.T) {
	m := NewManager()
	srv, wsURL := startWSTestServer(t)
	defer srv.Close()

	conn := mustDial(t, wsURL)
	defer func() {
		assert.NoError(t, conn.Close())
	}()
	assert.NoError(t, m.Register(10, conn))

	err := m.Send(&dto.NotificationResponse{UserID: 10, ActorID: 2, Type: "follow"})
	assert.NoError(t, err)
}

func TestManager_SendToDisconnectedUserReturnsError(t *testing.T) {
	m := NewManager()
	err := m.Send(&dto.NotificationResponse{UserID: 999, Type: "follow"})
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
		defer func() {
			_ = conn.Close()
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return nil
			}
		}
	})
	srv := httptest.NewServer(e)
	return srv, "ws" + srv.URL[len("http"):] + "/ws"
}

func mustDial(t *testing.T, wsURL string) *gws.Conn {
	t.Helper()
	conn, resp, err := gws.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		t.Fatalf("dial ws failed: %v (status=%d)", err, status)
	}
	return conn
}
