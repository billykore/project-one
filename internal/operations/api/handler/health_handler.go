package handler

import (
	"net/http"
	"time"

	"github.com/billykore/project-one/internal/operations/api/dto"
	operationsdomain "github.com/billykore/project-one/internal/operations/domain"
	operationsports "github.com/billykore/project-one/internal/operations/ports"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	"github.com/labstack/echo/v4"
)

// HealthHandler maps readiness assessments onto HTTP responses.
// It owns no dependency knowledge of its own; the use case assesses components.
type HealthHandler struct {
	healthUseCase operationsports.HealthAssessor
	log           platformports.Logger
}

// NewHealthHandler creates a new instance of HealthHandler.
func NewHealthHandler(healthUseCase operationsports.HealthAssessor, log platformports.Logger) *HealthHandler {
	return &HealthHandler{
		healthUseCase: healthUseCase,
		log:           log,
	}
}

// HandleLiveness handles the GET /healthz endpoint. It reports process liveness
// only and never contacts a dependency, so a running-but-not-ready process is
// still live.
//
//	@Summary		Liveness probe
//	@Description	Reports that the process can respond. It never contacts an external dependency and is not a readiness signal.
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	dto.LivenessResponse
//	@Router			/healthz [get]
func (h *HealthHandler) HandleLiveness(c echo.Context) error {
	return c.JSON(http.StatusOK, dto.LivenessResponse{
		Status:    "ok",
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
	})
}

// HandleReadiness handles the GET /status endpoint. It returns 503 whenever a
// required component is down, unknown, or could not be assessed in time, while
// still reporting the state of every assessed component. The report carries no
// address, credential, token, account identifier, or raw dependency error.
//
//	@Summary		Readiness probe
//	@Description	Reports whether the application can serve normal user traffic, plus the non-sensitive state and assessment time of every required dependency.
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	dto.HealthReportResponse	"Assessment is ready"
//	@Failure		503	{object}	dto.HealthReportResponse	"Assessment is not ready; the same report body is returned"
//	@Router			/status [get]
func (h *HealthHandler) HandleReadiness(c echo.Context) error {
	assessment := h.healthUseCase.Assess(c.Request().Context())

	status := http.StatusOK
	if !assessment.Ready() {
		status = http.StatusServiceUnavailable
		if h.log != nil {
			h.log.Warn(c.Request().Context(), "readiness assessment is not ready",
				"checked_at", assessment.CheckedAt.Format(time.RFC3339),
				"components", len(assessment.Components),
			)
		}
	}

	return c.JSON(status, newHealthReportResponse(assessment))
}

func newHealthReportResponse(assessment operationsdomain.HealthAssessment) dto.HealthReportResponse {
	components := make([]dto.HealthComponentResponse, 0, len(assessment.Components))
	for _, component := range assessment.Components {
		components = append(components, dto.HealthComponentResponse{
			Name:      component.Name,
			Status:    string(component.Status),
			CheckedAt: component.CheckedAt.UTC().Format(time.RFC3339),
		})
	}

	return dto.HealthReportResponse{
		Status:     string(assessment.Status),
		CheckedAt:  assessment.CheckedAt.UTC().Format(time.RFC3339),
		Components: components,
	}
}
