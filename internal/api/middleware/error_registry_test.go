package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestLookupError_Known(t *testing.T) {
	tests := []struct {
		err      error
		wantCode int
		wantStr  string
	}{
		{domain.ErrUserNotFound, http.StatusNotFound, domain.CodeNotFound},
		{domain.ErrInvalidCredentials, http.StatusUnauthorized, domain.CodeUnauthenticated},
		{domain.ErrUnauthorized, http.StatusUnauthorized, domain.CodeUnauthenticated},
		{domain.ErrInternalServer, http.StatusInternalServerError, domain.CodeInternal},
		{domain.ErrValidationFailed, http.StatusBadRequest, domain.CodeInvalidArgument},
		{domain.ErrAlreadyFollowing, http.StatusConflict, domain.CodeAlreadyExists},
		{domain.ErrCannotFollowSelf, http.StatusUnprocessableEntity, domain.CodeInvalidArgument},
		{domain.ErrPostNotFound, http.StatusNotFound, domain.CodeNotFound},
		{domain.ErrEmailAlreadyRegistered, http.StatusConflict, domain.CodeAlreadyExists},
		{domain.ErrNotFollowing, http.StatusNotFound, domain.CodeNotFound},
		{domain.ErrCannotUnfollowSelf, http.StatusUnprocessableEntity, domain.CodeInvalidArgument},
		{domain.ErrUsernameAlreadyTaken, http.StatusConflict, domain.CodeAlreadyExists},
		{domain.ErrInvalidPost, http.StatusBadRequest, domain.CodeInvalidArgument},
		{domain.ErrAlreadyLiked, http.StatusConflict, domain.CodeAlreadyExists},
		{domain.ErrNotLiked, http.StatusNotFound, domain.CodeNotFound},
		{domain.ErrCommentNotFound, http.StatusNotFound, domain.CodeNotFound},
		{domain.ErrNotificationNotFound, http.StatusNotFound, domain.CodeNotFound},
		{domain.ErrInvalidNotification, http.StatusBadRequest, domain.CodeInvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.wantStr, func(t *testing.T) {
			m := LookupError(tt.err)
			assert.Equal(t, tt.wantCode, m.Status)
			assert.Equal(t, tt.wantStr, m.Code)
			assert.NotEmpty(t, m.Detail)
			assert.NotEmpty(t, m.Title)
		})
	}
}

func TestLookupError_Unknown(t *testing.T) {
	m := LookupError(errors.New("some random error"))
	assert.Equal(t, http.StatusInternalServerError, m.Status)
	assert.Equal(t, domain.CodeInternal, m.Code)
	assert.Equal(t, "Internal Server Error", m.Title)
	assert.Equal(t, "Internal server error", m.Detail)
	assert.Empty(t, m.TypeSlug) // unknown → about:blank
}

func TestLookupError_Wrapped(t *testing.T) {
	// errors.Is should unwrap the chain
	wrapped := fmt.Errorf("get user: %w", domain.ErrUserNotFound)
	m := LookupError(wrapped)
	assert.Equal(t, http.StatusNotFound, m.Status)
	assert.Equal(t, domain.CodeNotFound, m.Code)
	assert.Equal(t, "not-found", m.TypeSlug)
	assert.Equal(t, "Not Found", m.Title)

	// double-wrapped
	doubleWrapped := fmt.Errorf("handler: %w", wrapped)
	m2 := LookupError(doubleWrapped)
	assert.Equal(t, http.StatusNotFound, m2.Status)
}
