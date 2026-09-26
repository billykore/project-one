// Package api exposes the HTTP routes owned by the publishing bounded context.
package api

import (
	featureflagmiddleware "github.com/billykore/project-one/internal/featureflags/api/middleware"
	featureflagports "github.com/billykore/project-one/internal/featureflags/ports"
	identitymiddleware "github.com/billykore/project-one/internal/identity/api/middleware"
	identityports "github.com/billykore/project-one/internal/identity/ports"
	publishinghandler "github.com/billykore/project-one/internal/publishing/api/handler"
	"github.com/labstack/echo/v4"
)

// RegisterRoutes mounts post, comment, and like endpoints.
func RegisterRoutes(
	e *echo.Echo,
	authenticator identityports.Authenticator,
	postCommandHandler *publishinghandler.PostCommandHandler,
	postQueryHandler *publishinghandler.PostQueryHandler,
	commentHandler *publishinghandler.CommentHandler,
	featureFlagEvaluator featureflagports.FeatureFlagEvaluator,
) {
	e.GET("/posts/:id", postQueryHandler.GetPostByID)

	posts := e.Group("/posts", identitymiddleware.Authorize(authenticator))
	posts.POST("", postCommandHandler.CreatePost, featureflagmiddleware.FeatureFlagGate(featureFlagEvaluator, "post-creation"))
	posts.GET("", postQueryHandler.GetPosts)
	posts.PUT("/:id", postCommandHandler.UpdatePost)
	posts.DELETE("/:id", postCommandHandler.DeletePost)
	posts.POST("/:id/comments", postCommandHandler.CreateComment)
	posts.POST("/:id/likes", postCommandHandler.LikePost)
	posts.DELETE("/:id/likes", postCommandHandler.UnlikePost)
	posts.GET("/:id/likes", postQueryHandler.GetLikeStatus)

	comments := e.Group("/comments", identitymiddleware.Authorize(authenticator))
	comments.PUT("/:id", commentHandler.EditComment)
	comments.DELETE("/:id", commentHandler.DeleteComment)
}
