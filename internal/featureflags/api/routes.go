// Package api exposes the HTTP routes owned by the feature-flags bounded context.
package api

import (
	featureflaghandler "github.com/billykore/project-one/internal/featureflags/api/handler"
	featureflagmiddleware "github.com/billykore/project-one/internal/featureflags/api/middleware"
	identitymiddleware "github.com/billykore/project-one/internal/identity/api/middleware"
	identityports "github.com/billykore/project-one/internal/identity/ports"
	"github.com/labstack/echo/v4"
)

// RegisterRoutes mounts feature-flag evaluation and operator administration endpoints.
func RegisterRoutes(e *echo.Echo, authenticator identityports.Authenticator, featureFlagHandler *featureflaghandler.FeatureFlagHandler, operators []string) {
	featureFlags := e.Group(
		"/admin/feature-flags",
		identitymiddleware.Authorize(authenticator),
		featureflagmiddleware.OperatorOnly(operators),
	)
	featureFlags.GET("", featureFlagHandler.ListFlags)
	featureFlags.POST("", featureFlagHandler.CreateFlag)
	featureFlags.GET("/:key", featureFlagHandler.GetFlag)
	featureFlags.PUT("/:key", featureFlagHandler.UpdateFlag)
	featureFlags.PATCH("/:key/environment/:environment", featureFlagHandler.SetEnvironment)
	featureFlags.PUT("/:key/overrides", featureFlagHandler.SetOverrides)
	featureFlags.POST("/:key/archive", featureFlagHandler.Archive)
	featureFlags.GET("/:key/audit", featureFlagHandler.ListAudit)

	e.GET("/feature-flags/evaluate", featureFlagHandler.Evaluate, identitymiddleware.OptionalAuthorize(authenticator))
}
