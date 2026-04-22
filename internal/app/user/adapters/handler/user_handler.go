package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/billykore/project-one/internal/app/user/adapters/dto"
	"github.com/billykore/project-one/internal/app/user/core/domain"
	"github.com/billykore/project-one/internal/app/user/core/ports"
	"github.com/labstack/echo/v4"
)

type userHandler struct {
	svc ports.UserService
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(svc ports.UserService) *userHandler {
	return &userHandler{svc: svc}
}

func (h *userHandler) Me(c echo.Context) error {
	userID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
	}

	user, err := h.svc.GetCurrentUser(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal Server Error"})
	}

	res := dto.UserResponse{
		ID:    strconv.Itoa(user.ID),
		Email: user.Email,
	}

	return c.JSON(http.StatusOK, res)
}
