package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/usecase/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_HandleChangePassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	mockLoginUC := mocks.NewMockLoginUseCase(ctrl)
	mockFollowUC := mocks.NewMockFollowUseCase(ctrl)
	mockPostUC := mocks.NewMockPostUseCase(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	h := NewUserHandler(mockUserUC, mockLoginUC, mockFollowUC, mockPostUC, mockValidator, mockLogger)

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		reqBody := dto.ChangePasswordRequest{
			OldPassword: "oldpassword123",
			NewPassword: "newpassword123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/users/password", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("username", "testuser")

		mockValidator.EXPECT().Validate(reqBody).Return(nil)
		mockUserUC.EXPECT().
			ChangePassword(gomock.Any(), "testuser", "oldpassword123", "newpassword123").
			Return(nil)

		err := h.HandleChangePassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.MessageResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Password updated successfully", resp.Message)
	})

	t.Run("unauthorized - no user context", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/users/password", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.HandleChangePassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("bad request - invalid json", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/users/password", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("username", "testuser")

		err := h.HandleChangePassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("bad request - validation error", func(t *testing.T) {
		e := echo.New()
		reqBody := dto.ChangePasswordRequest{
			OldPassword: "short",
			NewPassword: "newpassword123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/users/password", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("username", "testuser")

		validationErr := errors.New("validation failed")
		mockValidator.EXPECT().Validate(reqBody).Return(validationErr)

		err := h.HandleChangePassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("unauthorized - invalid current password", func(t *testing.T) {
		e := echo.New()
		reqBody := dto.ChangePasswordRequest{
			OldPassword: "wrongpassword",
			NewPassword: "newpassword123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/users/password", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("username", "testuser")

		mockValidator.EXPECT().Validate(reqBody).Return(nil)
		mockUserUC.EXPECT().
			ChangePassword(gomock.Any(), "testuser", "wrongpassword", "newpassword123").
			Return(domain.ErrInvalidCredentials)

		err := h.HandleChangePassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		e := echo.New()
		reqBody := dto.ChangePasswordRequest{
			OldPassword: "oldpassword123",
			NewPassword: "newpassword123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/users/password", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("username", "testuser")

		mockValidator.EXPECT().Validate(reqBody).Return(nil)
		mockUserUC.EXPECT().
			ChangePassword(gomock.Any(), "testuser", "oldpassword123", "newpassword123").
			Return(errors.New("db error"))
		mockLogger.EXPECT().Error(gomock.Any(), "failed to change password", "username", "testuser", "error", gomock.Any())

		err := h.HandleChangePassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
