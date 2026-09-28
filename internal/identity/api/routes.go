// Package api exposes the HTTP routes owned by the identity bounded context.
package api

import (
	"github.com/billykore/project-one/internal/identity/api/handler"
	"github.com/billykore/project-one/internal/identity/api/middleware"
	"github.com/billykore/project-one/internal/identity/ports"
	"github.com/labstack/echo/v4"
)

// RegisterRoutes mounts identity and user-profile endpoints.
func RegisterRoutes(e *echo.Echo, authenticator ports.Authenticator, userHandler *handler.UserHandler) {
	auth := e.Group("/auth")
	auth.POST("/register", userHandler.HandleRegister)
	auth.POST("/login", userHandler.HandleLogin)
	auth.POST("/logout", userHandler.HandleLogout, middleware.Authorize(authenticator))

	users := e.Group("/users")
	users.GET("/search", userHandler.SearchUsers)
	users.GET("/:username", userHandler.GetUser)

	usersAuth := users.Group("", middleware.Authorize(authenticator))
	usersAuth.PUT("/password", userHandler.HandleChangePassword)
	usersAuth.PUT("/profile", userHandler.HandleUpdateProfile)
}
