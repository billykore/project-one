// Package api exposes the HTTP routes owned by the social bounded context.
package api

import (
	identitymiddleware "github.com/billykore/project-one/internal/identity/api/middleware"
	identityports "github.com/billykore/project-one/internal/identity/ports"
	socialhandler "github.com/billykore/project-one/internal/social/api/handler"
	"github.com/labstack/echo/v4"
)

// RegisterRoutes mounts authenticated personal-feed endpoints.
func RegisterRoutes(e *echo.Echo, authenticator identityports.Authenticator, feedHandler *socialhandler.FeedHandler) {
	feeds := e.Group("/feeds", identitymiddleware.Authorize(authenticator))
	feeds.GET("", feedHandler.HandleGetFeed)
}
