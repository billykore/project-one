package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	metricsadapter "github.com/billykore/project-one/internal/adapters/metrics"
	"github.com/labstack/echo/v4"
)

// newMetricsRouteForTest mounts the real metrics route with credentials read
// from a mounted secret file, mirroring the Compose deployment.
func newMetricsRouteForTest(t *testing.T, username, secret string) *echo.Echo {
	t.Helper()

	secretPath := filepath.Join(t.TempDir(), "metrics-password")
	if err := os.WriteFile(secretPath, []byte(secret), 0o600); err != nil {
		t.Fatal(err)
	}

	credential, err := metricsadapter.LoadCredential(username, secretPath)
	if err != nil {
		t.Fatalf("LoadCredential returned %v, want nil", err)
	}

	recorder := metricsadapter.NewPrometheus()
	recorder.ObserveRequest(context.Background(), "GET", "/posts/:id", "2xx")

	e := echo.New()
	registerMetricsRoute(e, recorder.Handler(credential))
	return e
}

func requestMetrics(t *testing.T, e *echo.Echo, username, password string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	if username != "" || password != "" {
		request.SetBasicAuth(username, password)
	}

	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)
	return response
}

func TestMetricsRouteRejectsRequestsWithoutCredentials(t *testing.T) {
	e := newMetricsRouteForTest(t, "projectone-metrics", "mounted-secret")

	anonymous := requestMetrics(t, e, "", "")
	if anonymous.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d for a request without credentials", anonymous.Code, http.StatusUnauthorized)
	}
	if strings.Contains(anonymous.Body.String(), "projectone_http_requests_total") {
		t.Fatalf("unauthenticated response leaked measurements: %s", anonymous.Body.String())
	}
}

func TestMetricsRouteRejectsInvalidCredentials(t *testing.T) {
	e := newMetricsRouteForTest(t, "projectone-metrics", "mounted-secret")

	tests := []struct {
		name     string
		username string
		password string
	}{
		{name: "wrong password", username: "projectone-metrics", password: "wrong-secret"},
		{name: "wrong username", username: "not-the-metrics-user", password: "mounted-secret"},
		{name: "empty password", username: "projectone-metrics", password: " "},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := requestMetrics(t, e, test.username, test.password)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
			if strings.Contains(response.Body.String(), "projectone_http_requests_total") {
				t.Fatalf("rejected response leaked measurements: %s", response.Body.String())
			}
		})
	}
}

func TestMetricsRouteServesMeasurementsWithValidCredentials(t *testing.T) {
	e := newMetricsRouteForTest(t, "projectone-metrics", "mounted-secret")

	response := requestMetrics(t, e, "projectone-metrics", "mounted-secret")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := response.Body.String()
	for _, expected := range []string{
		"projectone_http_requests_total",
		`route="/posts/:id"`,
		"process_start_time_seconds",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("exposition is missing %q:\n%s", expected, body)
		}
	}
}

func TestMetricsRouteRejectsEveryRequestWhenMonitoringIsUnconfigured(t *testing.T) {
	credential, err := metricsadapter.LoadCredential("", "")
	if err != nil || credential != nil {
		t.Fatalf("LoadCredential with monitoring disabled = (%v, %v), want (nil, nil)", credential, err)
	}

	e := echo.New()
	registerMetricsRoute(e, metricsadapter.NewPrometheus().Handler(credential))

	response := requestMetrics(t, e, "", "")
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}

	// Even a well-formed credential must not work when none is configured.
	response = requestMetrics(t, e, "projectone-metrics", "mounted-secret")
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d without configured monitoring credentials", response.Code, http.StatusUnauthorized)
	}
}
