package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/identity/api/handler"
	"github.com/billykore/project-one/internal/identity/domain"
	pkgdto "github.com/billykore/project-one/internal/pkg/api/dto"
	pkgmiddleware "github.com/billykore/project-one/internal/pkg/api/middleware"
	"github.com/billykore/project-one/internal/pkg/problem"
	"github.com/billykore/project-one/internal/testkit/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLoginRateLimitedReturns429WithRetryAfter(t *testing.T) {
	for _, tc := range []struct {
		name       string
		retryAfter time.Duration
		wantHeader string
	}{
		{name: "rounds partial seconds up", retryAfter: 41500 * time.Millisecond, wantHeader: "42"},
		{name: "never advertises zero", retryAfter: 0, wantHeader: "1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			loginUseCase := mocks.NewMockLoginUseCase(ctrl)
			validator := mocks.NewMockValidator(ctrl)
			logger := mocks.NewMockLogger(ctrl)
			logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			validator.EXPECT().Validate(gomock.Any()).Return(nil)
			loginUseCase.EXPECT().
				Login(gomock.Any(), "user@example.com", "password123").
				Return(nil, fmt.Errorf("login: %w", &domain.LoginThrottledError{RetryAfter: tc.retryAfter}))

			e := echo.New()
			e.HTTPErrorHandler = pkgmiddleware.ErrorHandler(logger, "http://localhost:8080/errors", false)
			RegisterRoutes(e, nil, handler.NewUserHandler(nil, loginUseCase, validator, logger, false))

			req := httptest.NewRequest(http.MethodPost, "/auth/login",
				strings.NewReader(`{"email":"user@example.com","password":"password123"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusTooManyRequests, rec.Code)
			assert.Equal(t, tc.wantHeader, rec.Header().Get("Retry-After"))

			var body pkgdto.ProblemDetail
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, problem.CodeRateLimited, body.Code)
			assert.Equal(t, "http://localhost:8080/errors/rate-limited", body.Type)
			assert.Equal(t, http.StatusTooManyRequests, body.Status)
		})
	}
}
