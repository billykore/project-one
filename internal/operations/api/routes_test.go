package api

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestRegisterRoutesMountsOperationsContract(t *testing.T) {
	e := echo.New()
	RegisterRoutes(e, nil, nil)

	require.ElementsMatch(t, []string{"GET /healthz", "GET /status", "GET /metrics"}, registeredRoutes(e))
}

func registeredRoutes(e *echo.Echo) []string {
	routes := e.Routes()
	result := make([]string, 0, len(routes))
	for _, route := range routes {
		result = append(result, route.Method+" "+route.Path)
	}
	return result
}
