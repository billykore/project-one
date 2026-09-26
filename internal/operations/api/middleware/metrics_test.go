package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	metricsadapter "github.com/billykore/project-one/internal/operations/adapters/metrics"
	platformmiddleware "github.com/billykore/project-one/internal/platform/api/middleware"
	"github.com/billykore/project-one/internal/testkit/mocks"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/mock/gomock"
)

// serve attaches the metrics middleware to a router and performs one request.
func serve(recorder *mocks.MockHTTPMetrics, method, target string, register func(e *echo.Echo)) int {
	e := echo.New()
	e.Use(Metrics(recorder))
	register(e)

	request := httptest.NewRequest(method, target, nil)
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)

	return response.Code
}

func okHandler(status int) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(status, "ok")
	}
}

func TestMetricsMiddlewareUsesTheResolvedRouteTemplate(t *testing.T) {
	controller := gomock.NewController(t)
	recorder := mocks.NewMockHTTPMetrics(controller)
	recorder.EXPECT().ObserveRequest(gomock.Any(), "GET", "/posts/:id", "2xx").Times(1)
	recorder.EXPECT().ObserveRequestDuration(gomock.Any(), "GET", "/posts/:id", gomock.Any()).Times(1)

	status := serve(recorder, http.MethodGet, "/posts/42", func(e *echo.Echo) {
		e.GET("/posts/:id", okHandler(http.StatusOK))
	})

	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
}

func TestMetricsMiddlewareRecordsTheStatusClassTheErrorHandlerWillWrite(t *testing.T) {
	controller := gomock.NewController(t)
	recorder := mocks.NewMockHTTPMetrics(controller)
	recorder.EXPECT().ObserveRequest(gomock.Any(), "GET", "/missing", "4xx").Times(1)
	recorder.EXPECT().ObserveRequestDuration(gomock.Any(), "GET", "/missing", gomock.Any()).Times(1)
	recorder.EXPECT().ObserveRequest(gomock.Any(), "GET", "/boom", "5xx").Times(1)
	recorder.EXPECT().ObserveRequestDuration(gomock.Any(), "GET", "/boom", gomock.Any()).Times(1)

	serve(recorder, http.MethodGet, "/missing", func(e *echo.Echo) {
		e.GET("/missing", func(c echo.Context) error { return echo.ErrNotFound })
	})
	serve(recorder, http.MethodGet, "/boom", func(e *echo.Echo) {
		e.GET("/boom", func(c echo.Context) error { return errors.New("dependency exploded") })
	})
}

func TestMetricsMiddlewareLabelsEveryUnmatchedRouteAsOneSeries(t *testing.T) {
	controller := gomock.NewController(t)
	recorder := mocks.NewMockHTTPMetrics(controller)
	recorder.EXPECT().ObserveRequest(gomock.Any(), "GET", UnmatchedRoute, "4xx").Times(2)
	recorder.EXPECT().ObserveRequestDuration(gomock.Any(), "GET", UnmatchedRoute, gomock.Any()).Times(2)

	serve(recorder, http.MethodGet, "/no/such/route", func(*echo.Echo) {})
	serve(recorder, http.MethodGet, "/another/unknown/route?token=secret", func(*echo.Echo) {})
}

func TestMetricsMiddlewareExcludesItsOwnScrape(t *testing.T) {
	controller := gomock.NewController(t)
	// No EXPECT: observing the scrape route itself fails the test.
	recorder := mocks.NewMockHTTPMetrics(controller)

	status := serve(recorder, http.MethodGet, SelfScrapeRoute, func(e *echo.Echo) {
		e.GET(SelfScrapeRoute, okHandler(http.StatusOK))
	})

	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
}

func TestMetricsMiddlewareRecordsSuccessfulWritesAs2xx(t *testing.T) {
	controller := gomock.NewController(t)
	recorder := mocks.NewMockHTTPMetrics(controller)
	recorder.EXPECT().ObserveRequest(gomock.Any(), "POST", "/posts", "2xx").Times(1)
	recorder.EXPECT().ObserveRequestDuration(gomock.Any(), "POST", "/posts", gomock.Any()).Times(1)

	serve(recorder, http.MethodPost, "/posts", func(e *echo.Echo) {
		e.POST("/posts", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })
	})
}

func TestStatusForErrorMatchesTheErrorHandlerMapping(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "no error", err: nil, want: http.StatusOK},
		{name: "not found", err: echo.ErrNotFound, want: http.StatusNotFound},
		{name: "unauthorized", err: echo.ErrUnauthorized, want: http.StatusUnauthorized},
		{name: "wrapped echo error", err: echo.NewHTTPError(http.StatusTeapot, "teapot"), want: http.StatusTeapot},
		{name: "unmapped error", err: errors.New("boom"), want: http.StatusInternalServerError},
		{name: "bad request", err: echo.ErrBadRequest, want: http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := platformmiddleware.StatusForError(test.err); got != test.want {
				t.Fatalf("StatusForError = %d, want %d", got, test.want)
			}
		})
	}
}

// expositionText flattens every metric name and label value an operator could
// read from a scrape, so a test can search it for sensitive values.
func expositionText(t *testing.T, gatherer prometheus.Gatherer) string {
	t.Helper()

	families, err := gatherer.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}

	var builder strings.Builder
	for _, family := range families {
		builder.WriteString(family.GetName())
		builder.WriteString("\n")
		for _, metric := range family.GetMetric() {
			for _, label := range metric.GetLabel() {
				builder.WriteString(label.GetName())
				builder.WriteString("=")
				builder.WriteString(label.GetValue())
				builder.WriteString("\n")
			}
		}
	}
	return builder.String()
}

func TestMetricsMiddlewareNeverLabelsUserProvidedValues(t *testing.T) {
	// The real recorder is used so the assertion covers the exposed measurements
	// rather than only the values handed to the port.
	recorder := metricsadapter.NewPrometheus()

	e := echo.New()
	e.Use(Metrics(recorder))
	e.Use(echomiddleware.RequestID())
	e.GET("/users/:username/posts/:id", okHandler(http.StatusOK))
	e.GET("/search", okHandler(http.StatusOK))
	e.GET("/notifications/stream", okHandler(http.StatusOK))

	// Every value below is representative user, session, or request content.
	sensitive := []string{
		"alice9f2c@example.com",
		"4111111111111111",
		"super-secret-token",
		"session-token-value",
		"request-id-abc123",
		"credit-card-in-query",
	}

	requests := []*http.Request{
		httptest.NewRequest(http.MethodGet, "/users/alice9f2c@example.com/posts/4111111111111111", nil),
		httptest.NewRequest(http.MethodGet, "/search?q=credit-card-in-query&token=super-secret-token", nil),
		httptest.NewRequest(http.MethodGet, "/notifications/stream", nil),
	}
	requests[2].Header.Set("Authorization", "Bearer super-secret-token")
	requests[2].Header.Set("Cookie", "access_token=session-token-value")
	requests[2].Header.Set(echo.HeaderXRequestID, "request-id-abc123")

	for _, request := range requests {
		e.ServeHTTP(httptest.NewRecorder(), request)
	}

	exposition := expositionText(t, recorder.Gatherer())

	// Positive control: the route templates really are measured.
	if !strings.Contains(exposition, "/users/:username/posts/:id") {
		t.Fatalf("route template was not observed; the privacy assertion would be vacuous:\n%s", exposition)
	}

	for _, value := range sensitive {
		if strings.Contains(exposition, value) {
			t.Fatalf("sensitive value %q leaked into the exposed measurements:\n%s", value, exposition)
		}
	}
}

func TestMetricsMiddlewareLabelsUnmatchedRequestsWithoutTheRawURL(t *testing.T) {
	recorder := metricsadapter.NewPrometheus()

	e := echo.New()
	e.Use(Metrics(recorder))

	e.ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/unknown/alice9f2c@example.com?token=super-secret-token", nil))

	exposition := expositionText(t, recorder.Gatherer())
	if !strings.Contains(exposition, UnmatchedRoute) {
		t.Fatalf("unmatched bucket was not observed:\n%s", exposition)
	}
	for _, value := range []string{"alice9f2c@example.com", "super-secret-token"} {
		if strings.Contains(exposition, value) {
			t.Fatalf("sensitive value %q leaked from an unmatched request:\n%s", value, exposition)
		}
	}
}
