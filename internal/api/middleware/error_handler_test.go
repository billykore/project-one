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

	handler := ErrorHandler(&stubLogger{}, "", false)
	handler(domain.ErrUserNotFound, c)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	// RFC 9457: no {"error": {...}} wrapper — fields at top level
	assert.Equal(t, "http://localhost:8080/errors/not-found", body["type"])
	assert.Equal(t, "Not Found", body["title"])
	assert.Equal(t, float64(404), body["status"])
	assert.Equal(t, "User not found", body["detail"])
	assert.Equal(t, "/users/nonexistent", body["instance"])
	assert.Equal(t, "NOT_FOUND", body["code"])
	assert.Equal(t, "req_test_001", body["request_id"])
}

func TestErrorHandler_UnknownDefaultsTo500(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "req_test_002")

	handler := ErrorHandler(&stubLogger{}, "", true) // production mode
	handler(assert.AnError, c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "about:blank", body["type"])
	assert.Equal(t, "Internal Server Error", body["title"])
	assert.Equal(t, float64(500), body["status"])
	// In production, detail must be the generic default, not assert.AnError's message
	assert.Equal(t, "Internal server error", body["detail"])
	assert.Equal(t, "INTERNAL", body["code"])
}

func TestErrorHandler_HTTPErrorPassthrough(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "req_test_003")

	handler := ErrorHandler(&stubLogger{}, "", false)
	handler(echo.NewHTTPError(http.StatusUnprocessableEntity, "custom status"), c)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
}

func TestErrorHandler_WrappedError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "req_test_004")

	handler := ErrorHandler(&stubLogger{}, "", false)
	handler(echo.NewHTTPError(http.StatusBadRequest, "wraps validation").SetInternal(domain.ErrInvalidUser), c)

	assert.Equal(t, http.StatusBadRequest, rec.Code) // HTTPError code takes priority
}

func TestErrorHandler_ValidationErrors(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/auth/register", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "req_test_005")

	handler := ErrorHandler(&stubLogger{}, "", false)
	handler(echo.ErrBadRequest, c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "http://localhost:8080/errors/invalid-argument", body["type"])
	assert.Equal(t, "Bad Request", body["title"])
	assert.Equal(t, float64(400), body["status"])
	assert.Equal(t, "INVALID_ARGUMENT", body["code"])
	// errors extension is absent when no actual validator errors exist (omitempty)
	_, hasErrors := body["errors"]
	assert.False(t, hasErrors)
}
