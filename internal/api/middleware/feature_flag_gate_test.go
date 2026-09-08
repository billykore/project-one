package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type gateEvaluator struct{ decision domain.FeatureFlagDecision }

func (e gateEvaluator) Evaluate(context.Context, string, string) domain.FeatureFlagDecision {
	return e.decision
}

func (e gateEvaluator) Refresh(context.Context) error { return nil }

func TestFeatureFlagGateRejectsUnknownFlags(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/posts", nil)
	rec := httptest.NewRecorder()
	called := false
	next := FeatureFlagGate(gateEvaluator{decision: domain.FeatureFlagDecision{Source: domain.SourceUnknown}}, "post_creation")(
		func(echo.Context) error {
			called = true
			return nil
		},
	)

	err := next(e.NewContext(req, rec))

	assert.ErrorIs(t, err, echo.ErrNotFound)
	assert.False(t, called)
}
