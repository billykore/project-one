package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	sseadapter "github.com/billykore/project-one/internal/adapters/sse"
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
	userUc     ports.UserUseCase
	validator  ports.Validator
	sseManager *sseadapter.Manager
}

func NewNotificationHandler(
	log ports.Logger,
	subscriber ports.Subscriber,
	notificationUc ports.NotificationUseCase,
	userUc ports.UserUseCase,
	validator ports.Validator,
	sseManager *sseadapter.Manager,
) *NotificationHandler {
	// ponytail: nil checks removed — Go panics at method call site on nil pointer
	return &NotificationHandler{
		log:        log,
		subscriber: subscriber,
		uc:         notificationUc,
		userUc:     userUc,
		validator:  validator,
		sseManager: sseManager,
	}
}

// Listen starts a goroutine to listen for incoming notifications from the PubSub system
// and persists them to the database.
func (h *NotificationHandler) Listen(ctx context.Context) error {
	return h.subscriber.Subscribe(ctx, notificationTopic, func(ctx context.Context, event ports.Event) error {
		var notificationEvent domain.NotificationEvent
		if err := json.Unmarshal(event.Payload, &notificationEvent); err != nil {
			h.log.Error(ctx, "failed to unmarshal notification event", "error", err)
			return err
		}
		if notificationEvent.SchemaVersion != "" && notificationEvent.SchemaVersion != "1.0" {
			h.log.Warn(ctx, "unsupported notification schema version", "schema_version", notificationEvent.SchemaVersion)
			return fmt.Errorf("unsupported notification schema version: %s", notificationEvent.SchemaVersion)
		}

		notification := notificationEvent.Notification
		notification.EventID = notificationEvent.EventID

		if err := h.uc.SaveNotification(ctx, &notification); err != nil {
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
				h.log.Debug(ctx, "duplicate event skipped", "event_id", notification.EventID)
				return nil
			}
			h.log.Error(ctx, "failed to save notification", "error", err)
			return err
		}

		h.log.Info(ctx, "notification saved",
			"userID", notification.UserID,
			"actorID", notification.ActorID,
			"type", notification.Type,
		)

		resp := notificationResponseFromDomain(notification)

		var streamed bool
		err := h.sseManager.Send(resp)
		if err != nil {
			if !errors.Is(err, sseadapter.ErrUserNotConnected) {
				h.log.Warn(ctx, "failed to stream notification to sse", "userID", notification.UserID, "error", err)
			}
		} else {
			streamed = true
		}

		if streamed {
			h.log.Info(ctx, "notification streamed to realtime clients", "userID", notification.UserID, "type", notification.Type)
		}

		return nil
	})
}

// StreamNotifications handles the GET /notifications/stream endpoint.
//
//	@Summary		Stream notifications
//	@Description	Stream new notifications for the authenticated user with Server-Sent Events.
//	@Tags			notifications
//	@Produce		text/event-stream
//	@Success		200	{string}	string	"SSE stream opened"
//	@Failure		401	{object}	dto.ProblemDetail
//	@Failure		500	{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/notifications/stream [get]
func (h *NotificationHandler) StreamNotifications(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		h.log.Error(c.Request().Context(), "StreamNotifications failed", "error", "Username not found in context")
		return echo.ErrUnauthorized
	}

	user, err := h.userUc.GetUser(c.Request().Context(), username)
	if err != nil {
		h.log.Error(c.Request().Context(), "StreamNotifications failed", "username", username, "error", "User not found")
		return echo.ErrUnauthorized
	}

	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "streaming unsupported")
	}

	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
	c.Response().Header().Set(echo.HeaderConnection, "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")
	c.Response().WriteHeader(http.StatusOK)
	flusher.Flush()

	msgCh := h.sseManager.Register(user.ID)
	defer h.sseManager.Unregister(user.ID, msgCh)

	keepAlive := time.NewTicker(30 * time.Second)
	defer keepAlive.Stop()

	ctx := c.Request().Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-keepAlive.C:
			if _, err := c.Response().Write([]byte(": keepalive\n\n")); err != nil {
				h.log.Error(ctx, "failed to write keepalive", "userID", user.ID, "error", err)
				return nil
			}
			flusher.Flush()
		case msg, ok := <-msgCh:
			if !ok {
				return nil
			}
			payload, err := json.Marshal(msg)
			if err != nil {
				h.log.Warn(ctx, "failed to marshal sse notification", "userID", user.ID, "error", err)
				continue
			}
			if _, err := fmt.Fprintf(c.Response(), "event: notification\ndata: %s\n\n", payload); err != nil {
				return nil
			}
			flusher.Flush()
		}
	}
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
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/notifications [get]
func (h *NotificationHandler) GetNotifications(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		h.log.Error(c.Request().Context(), "GetNotifications failed", "error", "Username not found in context")
		return echo.ErrUnauthorized
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
		h.log.Error(c.Request().Context(), "GetNotifications failed", "username", username, "error", err)
		return err
	}

	resp := make([]*dto.NotificationResponse, len(notifications))
	for i, n := range notifications {
		resp[i] = notificationResponseFromDetail(n)
	}

	h.log.Info(c.Request().Context(), "GetNotifications succeeded", "username", username, "count", len(resp))
	return c.JSON(http.StatusOK, resp)
}

// MarkAsRead handles the PUT /notifications/:id/read endpoint.
//
//	@Summary		Mark notification as read
//	@Description	Mark a specific notification as read.
//	@Tags			notifications
//	@Param			id	path		int	true	"Notification ID"
//	@Success		200	{object}	dto.MessageResponse
//	@Failure		400	{object}	dto.ProblemDetail
//	@Failure		401	{object}	dto.ProblemDetail
//	@Failure		403	{object}	dto.ProblemDetail
//	@Failure		404	{object}	dto.ProblemDetail
//	@Failure		500	{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		h.log.Error(c.Request().Context(), "MarkAsRead failed", "error", "Username not found in context")
		return echo.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		h.log.Error(c.Request().Context(), "MarkAsRead failed", "username", username, "error", "Invalid notification ID")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid notification ID")
	}

	err = h.uc.MarkAsRead(c.Request().Context(), id, username)
	if err != nil {
		h.log.Error(c.Request().Context(), "MarkAsRead failed", "username", username, "notification_id", id, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "MarkAsRead succeeded", "username", username, "notification_id", id)
	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "Notification marked as read"})
}

// MarkAllAsRead handles the PUT /notifications/read-all endpoint.
//
//	@Summary		Mark all notifications as read
//	@Description	Mark all notifications for the authenticated user as read.
//	@Tags			notifications
//	@Success		200	{object}	dto.MessageResponse
//	@Failure		401	{object}	dto.ProblemDetail
//	@Failure		500	{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		h.log.Error(c.Request().Context(), "MarkAllAsRead failed", "error", "Username not found in context")
		return echo.ErrUnauthorized
	}

	err := h.uc.MarkAllAsRead(c.Request().Context(), username)
	if err != nil {
		h.log.Error(c.Request().Context(), "MarkAllAsRead failed", "username", username, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "MarkAllAsRead succeeded", "username", username)
	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "All notifications marked as read"})
}

func notificationResponseFromDomain(notification domain.Notification) *dto.NotificationResponse {
	return &dto.NotificationResponse{
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
	}
}

func notificationResponseFromDetail(notification *domain.NotificationDetail) *dto.NotificationResponse {
	return &dto.NotificationResponse{
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
	}
}
