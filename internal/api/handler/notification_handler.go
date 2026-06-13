package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/labstack/echo/v4"
)

type NotificationHandler struct {
	log       ports.Logger
	consumer  ports.NotificationConsumer
	uc        ports.NotificationUseCase
	validator ports.Validator
}

func NewNotificationHandler(
	log ports.Logger,
	consumer ports.NotificationConsumer,
	notificationUc ports.NotificationUseCase,
	validator ports.Validator,
) *NotificationHandler {
	if log == nil {
		panic("log is required")
	}
	if consumer == nil {
		panic("consumer is required")
	}
	if notificationUc == nil {
		panic("notificationUc is required")
	}
	if validator == nil {
		panic("validator is required")
	}
	return &NotificationHandler{
		log:       log,
		consumer:  consumer,
		uc:        notificationUc,
		validator: validator,
	}
}

// Consume starts the notification consumer to listen for incoming notifications and persist them.
// This should be called once during application startup and not called as a regular API endpoint.
func (h *NotificationHandler) Consume(ctx context.Context) error {
	outCh, err := h.consumer.Start(ctx)
	if err != nil {
		return err
	}

	go func(ctx context.Context) {
		for n := range outCh {
			if n == nil {
				continue
			}
			if err := h.uc.SaveNotification(ctx, n); err != nil {
				h.log.Error(ctx, "failed to persist notification", "userID", n.UserID, "type", n.Type, "error", err)
			}
			h.log.Info(ctx, "notification persisted successfully", "id", n.ID, "userID", n.UserID)
		}
	}(ctx)

	return nil
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
			ID:        n.ID,
			UserID:    n.UserID,
			ActorID:   n.ActorID,
			ActorName: n.ActorUsername,
			Type:      string(n.Type),
			PostID:    n.PostID,
			CommentID: n.CommentID,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt,
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
