package api

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestRegisterRoutesMountsSocialContract(t *testing.T) {
	e := echo.New()
	RegisterRoutes(e, nil, nil, nil)

	require.ElementsMatch(t, []string{
		"GET /feeds",
		"GET /users/:username/following", "GET /users/:username/followers",
		"POST /users/:username/followers", "DELETE /users/:username/followers",
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
