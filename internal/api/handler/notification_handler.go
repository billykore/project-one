package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	wsadapter "github.com/billykore/project-one/internal/adapters/websocket"
	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/labstack/echo/v4"
)

const notificationTopic = "notifications"

type NotificationHandler struct {
	log        ports.Logger
	subscriber ports.Subscriber
	uc         ports.NotificationUseCase
	validator  ports.Validator
	wsManager  *wsadapter.Manager
}

func NewNotificationHandler(
	log ports.Logger,
	subscriber ports.Subscriber,
	notificationUc ports.NotificationUseCase,
	validator ports.Validator,
	wsManager *wsadapter.Manager,
) *NotificationHandler {
	if log == nil {
		panic("log is required")
	}
	if subscriber == nil {
		panic("subscriber is required")
	}
	if notificationUc == nil {
		panic("notificationUc is required")
	}
	if validator == nil {
		panic("validator is required")
	}
	if wsManager == nil {
		panic("wsManager is required")
	}
	return &NotificationHandler{
		log:        log,
		subscriber: subscriber,
		uc:         notificationUc,
		validator:  validator,
		wsManager:  wsManager,
	}
}

// Listen starts a goroutine to listen for incoming notifications from the PubSub system
// and persists them to the database.
func (h *NotificationHandler) Listen(ctx context.Context) error {
	return h.subscriber.Subscribe(ctx, notificationTopic, func(ctx context.Context, event ports.Event) error {
		var notification domain.Notification
		if err := json.Unmarshal(event.Payload, &notification); err != nil {
			h.log.Error(ctx, "failed to unmarshal notification event", "error", err)
			return nil
		}

		if err := notification.Validate(); err != nil {
			h.log.Error(ctx, "invalid notification event", "error", err)
			return nil
		}

		if err := h.uc.SaveNotification(ctx, &notification); err != nil {
			h.log.Error(ctx, "failed to save notification", "error", err)
			return err
		}

		h.log.Info(ctx, "notification saved",
			"userID", notification.UserID,
			"actorID", notification.ActorID,
			"type", notification.Type,
		)

		if err := h.wsManager.Send(&dto.NotificationResponse{
			ID:            notification.ID,
			UserID:        notification.UserID,
			ActorID:       notification.ActorID,
			ActorUsername: notification.ActorUsername,
			Type:          string(notification.Type),
			PostID:        notification.PostID,
			CommentID:     notification.CommentID,
			IsRead:        notification.IsRead,
			CreatedAt:     notification.CreatedAt,
			Title:         dto.NotificationTitle(notification.Type),
			Body:          dto.NotificationBody(notification.Type, notification.ActorUsername),
		}); err != nil {
			h.log.Warn(ctx, "failed to stream notification to websocket", "userID", notification.UserID, "error", err)
			return nil
		}

		h.log.Info(ctx, "notification streamed to websocket", "userID", notification.UserID, "type", notification.Type)
		return nil
	})
}

// GetNotifications handles the GET /notifications endpoint.
//
//	@Summary		Get notifications
//	@Description	Retrieve notifications for the authenticated user.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int	false	"Limit"
//	@Param			offset	query		int	false	"Offset"
//	@Success		200		{array}		dto.NotificationResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/notifications [get]
func (h *NotificationHandler) GetNotifications(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit := 10
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = val
		}
	}
	offset := 0
	if offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			offset = val
		}
	}

	notifications, err := h.uc.GetNotifications(c.Request().Context(), username, limit, offset)
	if err != nil {
		h.log.Error(c.Request().Context(), "failed to get notifications", "username", username, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal server error"})
	}

	resp := make([]dto.NotificationResponse, len(notifications))
	for i, n := range notifications {
		resp[i] = dto.NotificationResponse{
			ID:            n.ID,
			UserID:        n.UserID,
			ActorID:       n.ActorID,
			ActorUsername: n.ActorUsername,
			Type:          string(n.Type),
			PostID:        n.PostID,
			CommentID:     n.CommentID,
			IsRead:        n.IsRead,
			CreatedAt:     n.CreatedAt,
			Title:         dto.NotificationTitle(n.Type),
			Body:          dto.NotificationBody(n.Type, n.ActorUsername),
		}
	}

	return c.JSON(http.StatusOK, resp)
}

// MarkAsRead handles the PUT /notifications/:id/read endpoint.
//
//	@Summary		Mark notification as read
//	@Description	Mark a specific notification as read.
//	@Tags			notifications
//	@Param			id	path		int	true	"Notification ID"
//	@Success		200	{object}	dto.MessageResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid notification ID"})
	}

	err = h.uc.MarkAsRead(c.Request().Context(), id, username)
	if err != nil {
		if errors.Is(err, domain.ErrNotificationNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Notification not found"})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "Forbidden"})
		}
		h.log.Error(c.Request().Context(), "failed to mark notification as read", "id", id, "username", username, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal server error"})
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Notification marked as read"})
}

// MarkAllAsRead handles the PUT /notifications/read-all endpoint.
//
//	@Summary		Mark all notifications as read
//	@Description	Mark all notifications for the authenticated user as read.
//	@Tags			notifications
//	@Success		200	{object}	dto.MessageResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	err := h.uc.MarkAllAsRead(c.Request().Context(), username)
	if err != nil {
		h.log.Error(c.Request().Context(), "failed to mark all notifications as read", "username", username, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal server error"})
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "All notifications marked as read"})
}
