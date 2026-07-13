package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/labstack/echo/v4"
)

// stubLogger implements ports.Logger for testing.
type stubLogger struct{}

func (s *stubLogger) Debug(_ context.Context, _ string, _ ...any) {}
func (s *stubLogger) Info(_ context.Context, _ string, _ ...any)  {}
func (s *stubLogger) Warn(_ context.Context, _ string, _ ...any)  {}
func (s *stubLogger) Error(_ context.Context, _ string, _ ...any) {}
func (s *stubLogger) Fatal(_ context.Context, _ string, _ ...any) {}

func TestErrorHandler_MapsStatus(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "req_test_001")

	handler := ErrorHandler(&stubLogger{}, false)
	handler(domain.ErrUserNotFound, c)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	errObj := body["error"].(map[string]interface{})
	assert.Equal(t, "NOT_FOUND", errObj["code"])
	assert.Equal(t, "req_test_001", errObj["request_id"])
}

func TestErrorHandler_UnknownDefaultsTo500(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "req_test_002")

	handler := ErrorHandler(&stubLogger{}, true) // production mode
	handler(assert.AnError, c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	errObj := body["error"].(map[string]interface{})
	assert.Equal(t, "INTERNAL", errObj["code"])
	// In production, message must be the generic default, not assert.AnError's message
	assert.Equal(t, "Internal server error", errObj["message"])
}

func TestErrorHandler_HTTPErrorPassthrough(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "req_test_003")

	handler := ErrorHandler(&stubLogger{}, false)
	handler(echo.NewHTTPError(http.StatusUnprocessableEntity, "custom status"), c)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestErrorHandler_WrappedError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "req_test_004")

	handler := ErrorHandler(&stubLogger{}, false)
	handler(echo.NewHTTPError(http.StatusBadRequest, "wraps validation").SetInternal(domain.ErrValidationFailed), c)

	assert.Equal(t, http.StatusBadRequest, rec.Code) // HTTPError code takes priority
}
