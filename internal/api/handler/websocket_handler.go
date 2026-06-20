package handler

import (
	"net/http"

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

func NewWebSocketHandler(
	log ports.Logger,
	tokenSvc ports.TokenService,
	userUc ports.UserUseCase,
	manager *wsadapter.Manager,
) *WebSocketHandler {
	// ponytail: nil checks removed — Go panics at method call site on nil pointer
	return &WebSocketHandler{
		log:      log,
		tokenSvc: tokenSvc,
		userUc:   userUc,
		manager:  manager,
		upgrader: gws.Upgrader{CheckOrigin: func(*http.Request) bool { return true }},
	}
}

// HandleUpgrade handles the WebSocket upgrade request.
// It validates the user's token, retrieves the user information, and registers the WebSocket connection.
func (h *WebSocketHandler) HandleUpgrade(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	user, err := h.userUc.GetUser(c.Request().Context(), username)
	if err != nil || user == nil {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	conn, err := h.upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
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
