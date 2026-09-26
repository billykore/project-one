package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuthorize_RequiresAnActiveSessionForTheCurrentUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authenticator := mocks.NewMockAuthenticator(ctrl)

	e := echo.New()
	e.GET("/private", func(c echo.Context) error {
		user, ok := c.Get("user").(*domain.User)
		if !ok {
			return echo.ErrUnauthorized
		}
		return c.String(http.StatusOK, user.Username)
	}, Authorize(authenticator))

	t.Run("rejects a valid JWT that was revoked", func(t *testing.T) {
		authenticator.EXPECT().Authenticate(gomock.Any(), "revoked-token").Return(nil, domain.ErrUntrustedToken)

		req := httptest.NewRequest(http.MethodGet, "/private", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer revoked-token")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("uses the current identity instead of stale username claims", func(t *testing.T) {
		authenticator.EXPECT().Authenticate(gomock.Any(), "active-token").Return(&domain.User{ID: 7, Username: "new-name"}, nil)

		req := httptest.NewRequest(http.MethodGet, "/private", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer active-token")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "new-name", rec.Body.String())
	})
}
