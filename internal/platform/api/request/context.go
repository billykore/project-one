package request

import (
	"github.com/billykore/project-one/internal/identity/domain"
	"github.com/labstack/echo/v4"
)

func CurrentUser(c echo.Context) (*domain.User, bool) {
	user, ok := c.Get("user").(*domain.User)
	return user, ok && user != nil
}
