package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	operationsdomain "github.com/billykore/project-one/internal/operations/domain"
	operationsports "github.com/billykore/project-one/internal/operations/ports"
	healthusecase "github.com/billykore/project-one/internal/operations/usecase"
	"github.com/billykore/project-one/internal/testkit/mocks"
	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"
)

func newHealthRequest(t *testing.T, target string) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, target, nil)
	response := httptest.NewRecorder()
	return echo.New().NewContext(request, response), response
}

func decodeHealthBody(t *testing.T, response *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not valid JSON: %v (%s)", err, response.Body.String())
	}
	return body
}

func assertRFC3339UTC(t *testing.T, value any, field string) {
	t.Helper()

	raw, ok := value.(string)
	if !ok {
		t.Fatalf("%s = %v, want an RFC 3339 string", field, value)
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatalf("%s = %q, want an RFC 3339 timestamp: %v", field, raw, err)
	}
	if parsed.Location() != time.UTC {
		t.Fatalf("%s = %q, want a UTC timestamp", field, raw)
	}
}

func healthyAssessment() operationsdomain.HealthAssessment {
	checkedAt := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	return operationsdomain.HealthAssessment{
		Status:    operationsdomain.ReadinessStatusReady,
		CheckedAt: checkedAt,
		Components: []operationsdomain.ComponentHealth{
			{Name: "database", Status: operationsdomain.ComponentStatusUp, CheckedAt: checkedAt},
			{Name: "rabbitmq", Status: operationsdomain.ComponentStatusUp, CheckedAt: checkedAt},
		},
	}
}

func TestHealthHandlerLivenessDoesNotContactDependencies(t *testing.T) {
	controller := gomock.NewController(t)
	// No EXPECT: any dependency assessment during liveness fails the test.
	assessor := mocks.NewMockHealthAssessor(controller)
	healthHandler := NewHealthHandler(assessor, nil)

	context, response := newHealthRequest(t, "/healthz")
	if err := healthHandler.HandleLiveness(context); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := decodeHealthBody(t, response)
	if body["status"] != "ok" {
		t.Fatalf("status = %v, want ok", body["status"])
	}
	assertRFC3339UTC(t, body["checked_at"], "checked_at")
}

func TestHealthHandlerReadinessReportsReadyComponents(t *testing.T) {
	controller := gomock.NewController(t)
	assessor := mocks.NewMockHealthAssessor(controller)
	assessor.EXPECT().Assess(gomock.Any()).Return(healthyAssessment())
	healthHandler := NewHealthHandler(assessor, nil)

	context, response := newHealthRequest(t, "/status")
	if err := healthHandler.HandleReadiness(context); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := decodeHealthBody(t, response)
	if body["status"] != "ready" {
		t.Fatalf("status = %v, want ready", body["status"])
	}
	assertRFC3339UTC(t, body["checked_at"], "checked_at")

	components, ok := body["components"].([]any)
	if !ok || len(components) != 2 {
		t.Fatalf("components = %v, want the database and broker entries", body["components"])
	}
	first, ok := components[0].(map[string]any)
	if !ok {
		t.Fatalf("component = %v, want an object", components[0])
	}
	if first["name"] != "database" || first["status"] != "up" {
		t.Fatalf("component = %v, want database up", first)
	}
	assertRFC3339UTC(t, first["checked_at"], "components[0].checked_at")
}

func TestHealthHandlerReadinessReturns503ForAFailedComponent(t *testing.T) {
	controller := gomock.NewController(t)
	assessor := mocks.NewMockHealthAssessor(controller)
	assessment := healthyAssessment()
	assessment.Status = operationsdomain.ReadinessStatusNotReady
	assessment.Components[0].Status = operationsdomain.ComponentStatusDown
	assessor.EXPECT().Assess(gomock.Any()).Return(assessment)
	healthHandler := NewHealthHandler(assessor, nil)

	context, response := newHealthRequest(t, "/status")
	if err := healthHandler.HandleReadiness(context); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}

	body := decodeHealthBody(t, response)
	if body["status"] != "not_ready" {
		t.Fatalf("status = %v, want not_ready", body["status"])
	}
	components, ok := body["components"].([]any)
	if !ok || len(components) != 2 {
		t.Fatalf("components = %v, want every assessed component preserved", body["components"])
	}
	failed, ok := components[0].(map[string]any)
	if !ok || failed["name"] != "database" || failed["status"] != "down" {
		t.Fatalf("component = %v, want database down", components[0])
	}
	healthy, ok := components[1].(map[string]any)
	if !ok || healthy["status"] != "up" {
		t.Fatalf("component = %v, want the healthy broker preserved", components[1])
	}
}

func TestHealthHandlerReadinessReturns503WhenAComponentIsUnknown(t *testing.T) {
	controller := gomock.NewController(t)
	assessor := mocks.NewMockHealthAssessor(controller)
	assessment := healthyAssessment()
	assessment.Status = operationsdomain.ReadinessStatusNotReady
	assessment.Components[1].Status = operationsdomain.ComponentStatusUnknown
	assessor.EXPECT().Assess(gomock.Any()).Return(assessment)
	healthHandler := NewHealthHandler(assessor, nil)

	context, response := newHealthRequest(t, "/status")
	if err := healthHandler.HandleReadiness(context); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}

	body := decodeHealthBody(t, response)
	components, ok := body["components"].([]any)
	if !ok || len(components) != 2 {
		t.Fatalf("components = %v, want the indeterminate component preserved", body["components"])
	}
	if component, ok := components[1].(map[string]any); !ok || component["status"] != "unknown" {
		t.Fatalf("component = %v, want rabbitmq unknown", components[1])
	}
}

// privacyCheckers returns dependency checkers whose failures carry the kind of
// detail an incident responder must never see in a health response.
type failingChecker struct {
	name string
	err  error
}

func (c failingChecker) Name() string                { return c.name }
func (c failingChecker) Check(context.Context) error { return c.err }

func TestHealthHandlerReadinessOmitsRawDependencyErrors(t *testing.T) {
	databaseErr := errors.New("dial tcp 10.1.2.3:5432: user=admin password=hunter2 database=project1: connection refused")
	brokerErr := errors.New("amqp://project1:broker-password@rabbitmq:5672/: dial tcp 10.1.2.9:5672: connection refused")

	useCase := healthusecase.NewHealthUseCase([]operationsports.DependencyChecker{
		failingChecker{name: "database", err: databaseErr},
		failingChecker{name: "rabbitmq", err: brokerErr},
	}, nil, time.Second, nil)
	healthHandler := NewHealthHandler(useCase, nil)

	context, response := newHealthRequest(t, "/status")
	if err := healthHandler.HandleReadiness(context); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}

	body := decodeHealthBody(t, response)
	if body["status"] != "not_ready" {
		t.Fatalf("status = %v, want not_ready", body["status"])
	}
	assertRFC3339UTC(t, body["checked_at"], "checked_at")

	components, ok := body["components"].([]any)
	if !ok || len(components) != 2 {
		t.Fatalf("components = %v, want both failed components identified", body["components"])
	}
	for index, want := range []map[string]string{{"name": "database", "status": "down"}, {"name": "rabbitmq", "status": "down"}} {
		component, ok := components[index].(map[string]any)
		if !ok {
			t.Fatalf("component = %v, want an object", components[index])
		}
		if component["name"] != want["name"] || component["status"] != want["status"] {
			t.Fatalf("component = %v, want %v", component, want)
		}
		assertRFC3339UTC(t, component["checked_at"], "components.checked_at")
	}

	// The state and time are present; every raw diagnostic value is not.
	raw := response.Body.String()
	for _, leaked := range []string{
		"10.1.2.3", "10.1.2.9", "5432", "5672", "admin", "hunter2",
		"project1", "broker-password", "connection refused", "dial tcp", "amqp://",
	} {
		if strings.Contains(raw, leaked) {
			t.Fatalf("health report leaked %q: %s", leaked, raw)
		}
	}
}
