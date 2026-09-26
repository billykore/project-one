// Package api exposes the HTTP routes owned by the notifications bounded context.
package api

import (
	identitymiddleware "github.com/billykore/project-one/internal/identity/api/middleware"
	identityports "github.com/billykore/project-one/internal/identity/ports"
	notificationhandler "github.com/billykore/project-one/internal/notifications/api/handler"
	"github.com/labstack/echo/v4"
)

// RegisterRoutes mounts authenticated notification history and SSE endpoints.
func RegisterRoutes(e *echo.Echo, authenticator identityports.Authenticator, notificationHandler *notificationhandler.NotificationHandler) {
	notifications := e.Group("/notifications", identitymiddleware.Authorize(authenticator))
	notifications.GET("", notificationHandler.GetNotifications)
	notifications.GET("/stream", notificationHandler.StreamNotifications)
	notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
	notifications.PUT("/read-all", notificationHandler.MarkAllAsRead)
}
