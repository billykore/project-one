package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/billykore/project-one/internal/featureflags/domain"
	"github.com/billykore/project-one/internal/testkit/mocks"
	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"
)

func TestFeatureFlagHandlerEvaluate(t *testing.T) {
	controller := gomock.NewController(t)
	useCase := mocks.NewMockFeatureFlagUseCase(controller)
	useCase.EXPECT().Evaluate(gomock.Any(), "editor", "").Return(domain.FeatureFlagDecision{Key: "editor", Enabled: true, Source: domain.SourceEnabledAll})

	handler := NewFeatureFlagHandler(useCase, nil, "local")
	e := echo.New()
	request := httptest.NewRequest("GET", "/feature-flags/evaluate?keys=editor", nil)
	response := httptest.NewRecorder()
	context := e.NewContext(request, response)
	if err := handler.Evaluate(context); err != nil {
		t.Fatal(err)
	}
	if response.Code != 200 || response.Body.String() == "" {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestFeatureFlagHandlerRejectsEmptyEvaluationKeys(t *testing.T) {
	controller := gomock.NewController(t)
	useCase := mocks.NewMockFeatureFlagUseCase(controller)
	handler := NewFeatureFlagHandler(useCase, nil, "local")
	e := echo.New()
	request := httptest.NewRequest("GET", "/feature-flags/evaluate", nil)
	response := httptest.NewRecorder()
	if err := handler.Evaluate(e.NewContext(request, response)); err != echo.ErrBadRequest {
		t.Fatalf("got %v, want %v", err, echo.ErrBadRequest)
	}
}
