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

func NewWebSocketHandler(
	log ports.Logger,
	tokenSvc ports.TokenService,
	userUc ports.UserUseCase,
	manager *wsadapter.Manager,
) *WebSocketHandler {
	if log == nil {
		panic("log is required")
	}
	if tokenSvc == nil {
		panic("tokenSvc is required")
	}
	if userUc == nil {
		panic("userUc is required")
	}
	if manager == nil {
		panic("manager is required")
	}

	return &WebSocketHandler{
		log:      log,
		tokenSvc: tokenSvc,
		userUc:   userUc,
		manager:  manager,
		upgrader: gws.Upgrader{CheckOrigin: func(*http.Request) bool { return true }},
	}
}

// HandleUpgrade handles the WebSocket upgrade request.
//
//	@Summary		Upgrade to WebSocket
//	@Description	Upgrade the HTTP connection to a WebSocket connection for real-time notifications.
//	@Tags			websocket
//	@Accept			json
//	@Produce		json
//	@Success		200		{array}		dto.NotificationResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/notifications [get]
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
