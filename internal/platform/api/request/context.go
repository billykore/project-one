package request

import (
	"github.com/billykore/project-one/internal/platform/auth"
	"github.com/labstack/echo/v4"
)

func CurrentUser(c echo.Context) (*auth.Principal, bool) {
	principal, ok := c.Get("principal").(*auth.Principal)
	if ok && principal != nil {
		return principal, true
	}
	// Compatibility for handlers and tests still being migrated from the old
	// context key. Authorization always writes principal.
	user, ok := c.Get("user").(*auth.Principal)
	return user, ok && user != nil
}
