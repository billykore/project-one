package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/usecase/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNotificationHandler_GetNotifications(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockNotificationUseCase(ctrl)
	mockConsumer := mocks.NewMockNotificationConsumer(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	hdl := NewNotificationHandler(mockLogger, mockConsumer, mockUseCase, mockValidator)

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/notifications?limit=5&offset=2", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("username", "user1")

		mockUseCase.EXPECT().
			GetNotifications(gomock.Any(), "user1", 5, 2).
			Return([]*domain.NotificationDetail{
				{
					Notification: domain.Notification{
						ID:     1,
						UserID: 1,
						Type:   domain.NotificationTypeFollow,
					},
					ActorUsername: "user2",
				},
			}, nil)

		err := hdl.GetNotifications(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("unauthorized - no user context", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := hdl.GetNotifications(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("username", "user1")

		mockUseCase.EXPECT().
			GetNotifications(gomock.Any(), "user1", 10, 0).
			Return(nil, errors.New("db error"))
		mockLogger.EXPECT().
			Error(gomock.Any(), "failed to get notifications", "username", "user1", "error", gomock.Any())

		err := hdl.GetNotifications(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestNotificationHandler_MarkAsRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockNotificationUseCase(ctrl)
	mockConsumer := mocks.NewMockNotificationConsumer(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	hdl := NewNotificationHandler(mockLogger, mockConsumer, mockUseCase, mockValidator)

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/notifications/1/read", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")
		c.Set("username", "user1")

		mockUseCase.EXPECT().
			MarkAsRead(gomock.Any(), 1, "user1").
			Return(nil)

		err := hdl.MarkAsRead(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("unauthorized - no user context", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/notifications/1/read", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := hdl.MarkAsRead(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("bad request - invalid ID", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/notifications/abc/read", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("abc")
		c.Set("username", "user1")

		err := hdl.MarkAsRead(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/notifications/99/read", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("99")
		c.Set("username", "user1")

		mockUseCase.EXPECT().
			MarkAsRead(gomock.Any(), 99, "user1").
			Return(domain.ErrNotificationNotFound)

		err := hdl.MarkAsRead(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/notifications/1/read", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")
		c.Set("username", "user1")

		mockUseCase.EXPECT().
			MarkAsRead(gomock.Any(), 1, "user1").
			Return(domain.ErrUnauthorized)

		err := hdl.MarkAsRead(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/notifications/1/read", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")
		c.Set("username", "user1")

		mockUseCase.EXPECT().
			MarkAsRead(gomock.Any(), 1, "user1").
			Return(errors.New("db error"))
		mockLogger.EXPECT().
			Error(gomock.Any(), "failed to mark notification as read", "id", 1, "username", "user1", "error", gomock.Any())

		err := hdl.MarkAsRead(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestNotificationHandler_MarkAllAsRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockNotificationUseCase(ctrl)
	mockConsumer := mocks.NewMockNotificationConsumer(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	hdl := NewNotificationHandler(mockLogger, mockConsumer, mockUseCase, mockValidator)

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/notifications/read-all", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("username", "user1")

		mockUseCase.EXPECT().
			MarkAllAsRead(gomock.Any(), "user1").
			Return(nil)

		err := hdl.MarkAllAsRead(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("unauthorized - no user context", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/notifications/read-all", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := hdl.MarkAllAsRead(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/notifications/read-all", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("username", "user1")

		mockUseCase.EXPECT().
			MarkAllAsRead(gomock.Any(), "user1").
			Return(errors.New("db error"))
		mockLogger.EXPECT().
			Error(gomock.Any(), "failed to mark all notifications as read", "username", "user1", "error", gomock.Any())

		err := hdl.MarkAllAsRead(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
