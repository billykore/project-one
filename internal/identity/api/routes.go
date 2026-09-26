// Package api exposes the HTTP routes owned by the identity bounded context.
package api

import (
	identityhandler "github.com/billykore/project-one/internal/identity/api/handler"
	identitymiddleware "github.com/billykore/project-one/internal/identity/api/middleware"
	identityports "github.com/billykore/project-one/internal/identity/ports"
	"github.com/labstack/echo/v4"
)

// RegisterRoutes mounts identity and user-profile endpoints.
func RegisterRoutes(e *echo.Echo, authenticator identityports.Authenticator, userHandler *identityhandler.UserHandler) {
	auth := e.Group("/auth")
	auth.POST("/register", userHandler.HandleRegister)
	auth.POST("/login", userHandler.HandleLogin)
	auth.POST("/logout", userHandler.HandleLogout, identitymiddleware.Authorize(authenticator))

	users := e.Group("/users")
	users.GET("/search", userHandler.SearchUsers)
	users.GET("/:username", userHandler.GetUser)
	users.GET("/:username/posts", userHandler.GetUserPosts)

	usersAuth := users.Group("", identitymiddleware.Authorize(authenticator))
	usersAuth.PUT("/password", userHandler.HandleChangePassword)
	usersAuth.PUT("/profile", userHandler.HandleUpdateProfile)
	usersAuth.GET("/:username/following", userHandler.GetFollowing)
	usersAuth.GET("/:username/followers", userHandler.GetFollowers)
	usersAuth.POST("/:username/followers", userHandler.HandleFollow)
	usersAuth.DELETE("/:username/followers", userHandler.HandleUnfollow)
}
