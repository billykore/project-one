package api

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestRegisterRoutesMountsIdentityContract(t *testing.T) {
	e := echo.New()
	RegisterRoutes(e, nil, nil)

	require.ElementsMatch(t, []string{
		"POST /auth/register", "POST /auth/login", "POST /auth/logout",
		"GET /users/search", "GET /users/:username",
		"PUT /users/password", "PUT /users/profile",
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
