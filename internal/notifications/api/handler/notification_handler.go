package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	sseadapter "github.com/billykore/project-one/internal/notifications/adapters/sse"
	"github.com/billykore/project-one/internal/notifications/api/dto"
	notificationdomain "github.com/billykore/project-one/internal/notifications/domain"
	notificationports "github.com/billykore/project-one/internal/notifications/ports"
	request "github.com/billykore/project-one/internal/pkg/api/request"
	vo "github.com/billykore/project-one/internal/pkg/pagination"
	pkgports "github.com/billykore/project-one/internal/pkg/ports"
	"github.com/billykore/project-one/internal/pkg/problem"
	"github.com/labstack/echo/v4"
)

type NotificationHandler struct {
	log        pkgports.Logger
	uc         notificationports.NotificationUseCase
	sseManager *sseadapter.Manager
}

func NewNotificationHandler(
	log pkgports.Logger,
	notificationUc notificationports.NotificationUseCase,
	sseManager *sseadapter.Manager,
) *NotificationHandler {
	// ponytail: nil checks removed — Go panics at method call site on nil pointer
	return &NotificationHandler{
		log:        log,
		uc:         notificationUc,
		sseManager: sseManager,
	}
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
	user, ok := request.CurrentUser(c)
	if !ok {
		h.log.Error(c.Request().Context(), "StreamNotifications failed", "error", "User not found in context")
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
//	@Param			cursor	query		string	false	"Pagination cursor from the previous response"
//	@Param			limit	query		int		false	"Items per page (1-100, default 10)"
//	@Success		200		{object}	dto.NotificationsListResponse
//	@Failure		401		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Security		BearerAuth
//	@Router			/notifications [get]
func (h *NotificationHandler) GetNotifications(c echo.Context) error {
	user, ok := request.CurrentUser(c)
	if !ok {
		h.log.Error(c.Request().Context(), "GetNotifications failed", "error", "User not found in context")
		return echo.ErrUnauthorized
	}

	limit := 10
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		value, err := strconv.Atoi(limitStr)
		if err != nil || value < 1 || value > 100 {
			return echo.ErrBadRequest
		}
		limit = value
	}
	var cursor *vo.Cursor
	if cursorStr := c.QueryParam("cursor"); cursorStr != "" {
		decoded, err := vo.DecodeCursor(cursorStr)
		if err != nil {
			return problem.ErrInvalidCursor
		}
		cursor = &decoded
	}

	notifications, err := h.uc.GetNotifications(c.Request().Context(), user, cursor, limit)
	if err != nil {
		h.log.Error(c.Request().Context(), "GetNotifications failed", "username", user.Username, "error", err)
		return err
	}

	resp := make([]dto.NotificationResponse, 0, len(notifications.Notifications))
	for _, n := range notifications.Notifications {
		resp = append(resp, *notificationResponseFromDetail(n))
	}

	h.log.Info(c.Request().Context(), "GetNotifications succeeded", "username", user.Username, "count", len(resp))
	nextCursor := ""
	if notifications.NextCursor != nil {
		nextCursor = notifications.NextCursor.Encode()
	}
	return c.JSON(http.StatusOK, dto.NotificationsListResponse{Data: resp, NextCursor: nextCursor, HasMore: notifications.HasMore})
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
	user, ok := request.CurrentUser(c)
	if !ok {
		h.log.Error(c.Request().Context(), "MarkAsRead failed", "error", "User not found in context")
		return echo.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		h.log.Error(c.Request().Context(), "MarkAsRead failed", "username", user.Username, "error", "Invalid notification ID")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid notification ID")
	}

	err = h.uc.MarkAsRead(c.Request().Context(), id, user.ID)
	if err != nil {
		h.log.Error(c.Request().Context(), "MarkAsRead failed", "username", user.Username, "notification_id", id, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "MarkAsRead succeeded", "username", user.Username, "notification_id", id)
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
	user, ok := request.CurrentUser(c)
	if !ok {
		h.log.Error(c.Request().Context(), "MarkAllAsRead failed", "error", "User not found in context")
		return echo.ErrUnauthorized
	}

	err := h.uc.MarkAllAsRead(c.Request().Context(), user.ID)
	if err != nil {
		h.log.Error(c.Request().Context(), "MarkAllAsRead failed", "username", user.Username, "error", err)
		return err
	}

	h.log.Info(c.Request().Context(), "MarkAllAsRead succeeded", "username", user.Username)
	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "All notifications marked as read"})
}

func notificationResponseFromDetail(notification *notificationdomain.NotificationDetail) *dto.NotificationResponse {
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
