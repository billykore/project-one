package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/billykore/project-one/internal/app/user/core/domain"
	"github.com/billykore/project-one/internal/app/user/core/service/mocks"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_Me(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockUserService(ctrl)
	mockLoginSvc := mocks.NewMockLoginService(ctrl)
	v := validator.New()
	h := NewUserHandler(mockSvc, mockLoginSvc, v)
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("userID", 1)

		user := &domain.User{FirstName: "John", LastName: "Doe", Email: "test@example.com"}
		mockSvc.EXPECT().GetCurrentUser(gomock.Any(), 1).Return(user, nil)

		if assert.NoError(t, h.Me(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var res UserResponse
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, "John Doe", res.Name)
			assert.Equal(t, "test@example.com", res.Email)
		}
	})

	t.Run("unauthorized_no_userID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, h.Me(c)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("internal_server_error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("userID", 1)

		mockSvc.EXPECT().GetCurrentUser(gomock.Any(), 1).Return(nil, errors.New("db error"))

		if assert.NoError(t, h.Me(c)) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestUserHandler_HandleLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockUserService(ctrl)
	mockLoginSvc := mocks.NewMockLoginService(ctrl)
	v := validator.New()
	h := NewUserHandler(mockSvc, mockLoginSvc, v)
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/user/login", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockLoginSvc.EXPECT().Login(gomock.Any(), reqBody.Email, reqBody.Password).
			Return("access", nil)

		if assert.NoError(t, h.HandleLogin(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// Assert cookies
			cookies := rec.Result().Cookies()
			assert.Len(t, cookies, 1)

			var accessTokenCookie *http.Cookie
			for _, cookie := range cookies {
				switch cookie.Name {
				case "access_token":
					accessTokenCookie = cookie
				}
			}

			assert.NotNil(t, accessTokenCookie)
			assert.Equal(t, "access", accessTokenCookie.Value)
			assert.True(t, accessTokenCookie.HttpOnly)

			var res LoginResponse
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, "Login successful", res.Message)
		}
	})

	t.Run("invalid_request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/user/login", bytes.NewBufferString("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, h.HandleLogin(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("validation_error", func(t *testing.T) {
		reqBody := LoginRequest{
			Email: "invalid-email",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/user/login", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, h.HandleLogin(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("invalid_credentials", func(t *testing.T) {
		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/user/login", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockLoginSvc.EXPECT().Login(gomock.Any(), reqBody.Email, reqBody.Password).
			Return("", domain.ErrInvalidCredentials)

		if assert.NoError(t, h.HandleLogin(c)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})
}

func TestUserHandler_HandleLogout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockUserService(ctrl)
	mockLoginSvc := mocks.NewMockLoginService(ctrl)
	v := validator.New()
	h := NewUserHandler(mockSvc, mockLoginSvc, v)
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/user/logout", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("userID", 1)

		mockLoginSvc.EXPECT().Logout(gomock.Any(), 1).Return(nil)

		if assert.NoError(t, h.HandleLogout(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var res LogoutResponse
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, "Logged out successfully", res.Message)
		}
	})

	t.Run("unauthorized_no_userID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/user/logout", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, h.HandleLogout(c)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("internal_server_error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/user/logout", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("userID", 1)

		mockLoginSvc.EXPECT().Logout(gomock.Any(), 1).Return(errors.New("logout error"))

		if assert.NoError(t, h.HandleLogout(c)) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestUserHandler_HandleRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockUserService(ctrl)
	mockLoginSvc := mocks.NewMockLoginService(ctrl)
	v := validator.New()
	h := NewUserHandler(mockSvc, mockLoginSvc, v)
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		reqBody := RegisterRequest{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Password:  "password123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/user/register", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockSvc.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil)

		if assert.NoError(t, h.HandleRegister(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
			var res RegisterResponse
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, "User registered successfully", res.Message)
		}
	})

	t.Run("duplicate_email", func(t *testing.T) {
		reqBody := RegisterRequest{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "exists@example.com",
			Password:  "password123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/user/register", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockSvc.EXPECT().Register(gomock.Any(), gomock.Any()).Return(domain.ErrEmailAlreadyRegistered)

		if assert.NoError(t, h.HandleRegister(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			var res ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, "Email is already registered", res.Error)
		}
	})

	t.Run("validation_error", func(t *testing.T) {
		reqBody := RegisterRequest{
			FirstName: "Jo", // too short
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/user/register", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, h.HandleRegister(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("internal_error", func(t *testing.T) {
		reqBody := RegisterRequest{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Password:  "password123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/user/register", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockSvc.EXPECT().Register(gomock.Any(), gomock.Any()).Return(errors.New("db error"))

		if assert.NoError(t, h.HandleRegister(c)) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			var res ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, "Something went wrong", res.Error)
		}
	})
}
