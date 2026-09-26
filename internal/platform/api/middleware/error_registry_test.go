package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/billykore/project-one/internal/platform/problem"
	publishingdomain "github.com/billykore/project-one/internal/publishing/domain"
	"github.com/stretchr/testify/assert"
)

func TestLookupError_Known(t *testing.T) {
	tests := []struct {
		err      error
		wantCode int
		wantStr  string
	}{
		{problem.ErrUserNotFound, http.StatusNotFound, problem.CodeNotFound},
		{problem.ErrInvalidCredentials, http.StatusUnauthorized, problem.CodeUnauthenticated},
		{problem.ErrAlreadyFollowing, http.StatusConflict, problem.CodeAlreadyExists},
		{problem.ErrCannotFollowSelf, http.StatusUnprocessableEntity, problem.CodeInvalidArgument},
		{problem.ErrPostNotFound, http.StatusNotFound, problem.CodeNotFound},
		{problem.ErrEmailAlreadyRegistered, http.StatusConflict, problem.CodeAlreadyExists},
		{problem.ErrNotFollowing, http.StatusNotFound, problem.CodeNotFound},
		{problem.ErrCannotUnfollowSelf, http.StatusUnprocessableEntity, problem.CodeInvalidArgument},
		{problem.ErrUsernameAlreadyTaken, http.StatusConflict, problem.CodeAlreadyExists},
		{problem.ErrInvalidPost, http.StatusBadRequest, problem.CodeInvalidArgument},
		{publishingdomain.ErrAlreadyLiked, http.StatusConflict, problem.CodeAlreadyExists},
		{publishingdomain.ErrNotLiked, http.StatusNotFound, problem.CodeNotFound},
		{problem.ErrCommentNotFound, http.StatusNotFound, problem.CodeNotFound},
		{problem.ErrNotificationNotFound, http.StatusNotFound, problem.CodeNotFound},
		{problem.ErrInvalidNotification, http.StatusBadRequest, problem.CodeInvalidArgument},
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
	assert.Equal(t, problem.CodeInternal, m.Code)
	assert.Equal(t, "Internal Server Error", m.Title)
	assert.Equal(t, "Something went wrong", m.Detail)
	assert.Empty(t, m.TypeSlug) // unknown → about:blank
}

func TestLookupError_Wrapped(t *testing.T) {
	// errors.Is should unwrap the chain
	wrapped := fmt.Errorf("get user: %w", problem.ErrUserNotFound)
	m := LookupError(wrapped)
	assert.Equal(t, http.StatusNotFound, m.Status)
	assert.Equal(t, problem.CodeNotFound, m.Code)
	assert.Equal(t, "not-found", m.TypeSlug)
	assert.Equal(t, "Not Found", m.Title)

	// double-wrapped
	doubleWrapped := fmt.Errorf("handler: %w", wrapped)
	m2 := LookupError(doubleWrapped)
	assert.Equal(t, http.StatusNotFound, m2.Status)
}
