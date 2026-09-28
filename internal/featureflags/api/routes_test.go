package api

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestRegisterRoutesMountsFeatureFlagContract(t *testing.T) {
	e := echo.New()
	RegisterRoutes(e, nil, nil, nil)

	require.ElementsMatch(t, []string{
		"GET /admin/feature-flags", "POST /admin/feature-flags",
		"GET /admin/feature-flags/:key", "PUT /admin/feature-flags/:key",
		"PATCH /admin/feature-flags/:key/environment/:environment",
		"PUT /admin/feature-flags/:key/overrides", "POST /admin/feature-flags/:key/archive",
		"GET /admin/feature-flags/:key/audit", "GET /feature-flags/evaluate",
	}, registeredRoutes(e))
}

func registeredRoutes(e *echo.Echo) []string {
	routes := e.Routes()
	result := make([]string, 0, len(routes))
	for _, route := range routes {
		if route.Method == "echo_route_not_found" {
			continue
		}
		result = append(result, route.Method+" "+route.Path)
	}
	return result
}
