package api

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestRegisterRoutesMountsPublishingContract(t *testing.T) {
	e := echo.New()
	RegisterRoutes(e, nil, nil, nil, nil, nil, nil)

	require.ElementsMatch(t, []string{
		"GET /posts/:id", "GET /users/:username/posts", "POST /posts", "GET /posts", "PUT /posts/:id", "DELETE /posts/:id",
		"POST /posts/:id/comments", "POST /posts/:id/likes", "DELETE /posts/:id/likes", "GET /posts/:id/likes",
		"PUT /comments/:id", "DELETE /comments/:id",
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
