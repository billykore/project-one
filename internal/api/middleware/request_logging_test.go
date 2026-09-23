package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

type capturedLog struct {
	level  string
	msg    string
	fields []any
}

type recordingLogger struct {
	entries []capturedLog
}

func (l *recordingLogger) Debug(_ context.Context, msg string, fields ...any) {
	l.capture("debug", msg, fields)
}

func (l *recordingLogger) Info(_ context.Context, msg string, fields ...any) {
	l.capture("info", msg, fields)
}

func (l *recordingLogger) Warn(_ context.Context, msg string, fields ...any) {
	l.capture("warn", msg, fields)
}

func (l *recordingLogger) Error(_ context.Context, msg string, fields ...any) {
	l.capture("error", msg, fields)
}

func (l *recordingLogger) Fatal(_ context.Context, msg string, fields ...any) {
	l.capture("fatal", msg, fields)
}

func (l *recordingLogger) capture(level, msg string, fields []any) {
	l.entries = append(l.entries, capturedLog{level: level, msg: msg, fields: fields})
}

func fieldsMap(t *testing.T, fields []any) map[string]any {
	t.Helper()
	if len(fields)%2 != 0 {
		t.Fatalf("fields length = %d, want key/value pairs", len(fields))
	}
	result := make(map[string]any, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			t.Fatalf("field key %d has type %T, want string", i, fields[i])
		}
		result[key] = fields[i+1]
	}
	return result
}

func TestRequestLoggingEmitsSafeCompletionFields(t *testing.T) {
	log := &recordingLogger{}
	e := echo.New()
	e.Use(RequestLogging(log))
	e.Use(echomiddleware.RequestID())
	e.GET("/posts/:id", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/posts/ACCOUNT_SENTINEL?token=TOKEN_SENTINEL", nil)
	request.Header.Set(echo.HeaderXRequestID, "request-safe-001")
	request.Header.Set("Authorization", "Bearer AUTH_SENTINEL")
	request.Header.Set("Cookie", "session=COOKIE_SENTINEL")
	request.Header.Set("User-Agent", "AGENT_SENTINEL")
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)

	if len(log.entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(log.entries))
	}
	entry := log.entries[0]
	if entry.level != "info" || entry.msg != RequestCompletedMessage {
		t.Fatalf("entry = (%q, %q), want (info, %q)", entry.level, entry.msg, RequestCompletedMessage)
	}

	fields := fieldsMap(t, entry.fields)
	want := map[string]any{
		"request_id": "request-safe-001",
		"method":     http.MethodGet,
		"route":      "/posts/:id",
		"status":     http.StatusNoContent,
	}
	for key, value := range want {
		if fields[key] != value {
			t.Errorf("field %q = %#v, want %#v", key, fields[key], value)
		}
	}
	if duration, ok := fields["duration_ms"].(float64); !ok || duration < 0 {
		t.Errorf("duration_ms = %#v, want non-negative float64", fields["duration_ms"])
	}

	serialized := strings.Builder{}
	serialized.WriteString(entry.msg)
	for key, value := range fields {
		serialized.WriteString(key)
		serialized.WriteString("=")
		serialized.WriteString(toString(value))
	}
	for _, prohibited := range []string{
		"ACCOUNT_SENTINEL", "TOKEN_SENTINEL", "AUTH_SENTINEL", "COOKIE_SENTINEL", "AGENT_SENTINEL",
		"raw_url", "query", "remote_addr", "user_agent", "authorization", "cookie", "body", "error", "stack_trace",
	} {
		if strings.Contains(strings.ToLower(serialized.String()), strings.ToLower(prohibited)) {
			t.Errorf("completion event contains prohibited value or field %q: %s", prohibited, serialized.String())
		}
	}
}

func TestRequestLoggingClassifiesFailuresWithoutRawErrors(t *testing.T) {
	log := &recordingLogger{}
	e := echo.New()
	e.Use(RequestLogging(log))
	e.Use(echomiddleware.RequestID())
	e.GET("/boom", func(echo.Context) error {
		return errors.New("PRIVATE_ERROR_SENTINEL")
	})

	request := httptest.NewRequest(http.MethodGet, "/boom", nil)
	request.Header.Set(echo.HeaderXRequestID, "request-error-001")
	e.ServeHTTP(httptest.NewRecorder(), request)

	if len(log.entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(log.entries))
	}
	entry := log.entries[0]
	fields := fieldsMap(t, entry.fields)
	if entry.level != "error" {
		t.Errorf("level = %q, want error", entry.level)
	}
	if fields["status"] != http.StatusInternalServerError {
		t.Errorf("status = %#v, want %d", fields["status"], http.StatusInternalServerError)
	}
	if fields["error_code"] != domain.CodeInternal {
		t.Errorf("error_code = %#v, want %q", fields["error_code"], domain.CodeInternal)
	}
	if fields["failure_category"] != FailureCategoryServer {
		t.Errorf("failure_category = %#v, want %q", fields["failure_category"], FailureCategoryServer)
	}
	for _, value := range entry.fields {
		if strings.Contains(toString(value), "PRIVATE_ERROR_SENTINEL") {
			t.Fatal("raw error leaked into completion event")
		}
	}
}

func toString(value any) string {
	return fmt.Sprint(value)
}
