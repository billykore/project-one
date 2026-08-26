package handler

import (
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/labstack/echo/v4"
)

func currentUser(c echo.Context) (*domain.User, bool) {
	user, ok := c.Get("user").(*domain.User)
	return user, ok && user != nil
}
