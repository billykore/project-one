package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/usecase/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	mockLoginUC := mocks.NewMockLoginUseCase(ctrl)
	mockFollowUC := mocks.NewMockFollowUseCase(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	h := NewUserHandler(mockUserUC, mockLoginUC, mockFollowUC, mockValidator, mockLogger)
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		username := "testuser"
		user := &domain.User{
			Username:  username,
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users/:username")
		c.SetParamNames("username")
		c.SetParamValues(username)

		mockUserUC.EXPECT().GetUserProfile(gomock.Any(), username).Return(user, nil)

		if assert.NoError(t, h.GetProfile(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var res dto.UserResponse
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, user.Username, res.Username)
			assert.Equal(t, user.Email, res.Email)
			assert.Equal(t, user.FirstName+" "+user.LastName, res.Name)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		username := "notfound"
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users/:username")
		c.SetParamNames("username")
		c.SetParamValues(username)

		mockUserUC.EXPECT().GetUserProfile(gomock.Any(), username).Return(nil, domain.ErrUserNotFound)

		if assert.NoError(t, h.GetProfile(c)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
			var res dto.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("User %s not found", username), res.Error)
		}
	})

	t.Run("invalid username", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users/:username")
		c.SetParamNames("username")
		c.SetParamValues("")

		if assert.NoError(t, h.GetProfile(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			var res dto.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, "Invalid username", res.Error)
		}
	})

	t.Run("internal server error", func(t *testing.T) {
		username := "testuser"
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users/:username")
		c.SetParamNames("username")
		c.SetParamValues(username)

		errSome := errors.New("something went wrong")
		mockUserUC.EXPECT().GetUserProfile(gomock.Any(), username).Return(nil, errSome)
		mockLogger.EXPECT().Error(gomock.Any(), "failed to get user profile", "username", username, "error", errSome)

		if assert.NoError(t, h.GetProfile(c)) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			var res dto.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, "Something went wrong", res.Error)
		}
	})
}
