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
	// ponytail: nil checks removed — Go panics at method call site on nil pointer
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

		if err := h.uc.SaveNotification(ctx, &notification); err != nil {
			h.log.Error(ctx, "failed to save notification", "error", err)
			return err
		}

		h.log.Info(ctx, "notification saved",
			"userID", notification.UserID,
			"actorID", notification.ActorID,
			"type", notification.Type,
		)

		err := h.wsManager.Send(&dto.NotificationResponse{
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
		})
		if err != nil {
			// User not being connected is a normal condition (user is offline);
			// only log actual send failures as warnings.
			if errors.Is(err, wsadapter.ErrUserNotConnected) {
				h.log.Debug(ctx, "user not connected to websocket, skipping stream", "userID", notification.UserID)
			} else {
				h.log.Warn(ctx, "failed to stream notification to websocket", "userID", notification.UserID, "error", err)
			}
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
//	@Failure		401		{object}	dto.APIErrorResponse
//	@Failure		500		{object}	dto.APIErrorResponse
//	@Security		BearerAuth
//	@Router			/notifications [get]
func (h *NotificationHandler) GetNotifications(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
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
		return err
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
//	@Failure		400	{object}	dto.APIErrorResponse
//	@Failure		401	{object}	dto.APIErrorResponse
//	@Failure		403	{object}	dto.APIErrorResponse
//	@Failure		404	{object}	dto.APIErrorResponse
//	@Failure		500	{object}	dto.APIErrorResponse
//	@Security		BearerAuth
//	@Router			/notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid notification ID")
	}

	err = h.uc.MarkAsRead(c.Request().Context(), id, username)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Notification marked as read"})
}

// MarkAllAsRead handles the PUT /notifications/read-all endpoint.
//
//	@Summary		Mark all notifications as read
//	@Description	Mark all notifications for the authenticated user as read.
//	@Tags			notifications
//	@Success		200	{object}	dto.MessageResponse
//	@Failure		401	{object}	dto.APIErrorResponse
//	@Failure		500	{object}	dto.APIErrorResponse
//	@Security		BearerAuth
//	@Router			/notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return domain.ErrUnauthorized
	}

	err := h.uc.MarkAllAsRead(c.Request().Context(), username)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "All notifications marked as read"})
}
