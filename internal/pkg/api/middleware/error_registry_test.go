package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/billykore/project-one/internal/pkg/problem"
	"github.com/stretchr/testify/assert"
)

func TestLookupError_Known(t *testing.T) {
	tests := []struct {
		err      error
		wantCode int
		wantStr  string
	}{
		{problem.ErrUserNotFound, http.StatusNotFound, problem.CodeNotFound},
		{problem.ErrInvalidCursor, http.StatusBadRequest, problem.CodeInvalidArgument},
		{problem.ErrInvalidArgument, http.StatusUnprocessableEntity, problem.CodeInvalidArgument},
		{problem.ErrFeatureDisabled, http.StatusNotFound, problem.CodeNotFound},
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

func TestRegisterErrorMappings(t *testing.T) {
	errContextOwned := errors.New("context-owned failure")
	mapping := ErrorMapping{Status: http.StatusConflict, Code: problem.CodeConflict, TypeSlug: "conflict", Title: "Conflict", Detail: "Context-owned failure"}

	RegisterErrorMappings(map[error]ErrorMapping{errContextOwned: mapping})
	assert.Equal(t, mapping, LookupError(fmt.Errorf("wrapped: %w", errContextOwned)))

	assert.NotPanics(t, func() {
		RegisterErrorMappings(map[error]ErrorMapping{errContextOwned: mapping})
	}, "re-registering an identical mapping is a no-op")

	conflicting := mapping
	conflicting.Status = http.StatusBadRequest
	assert.Panics(t, func() {
		RegisterErrorMappings(map[error]ErrorMapping{errContextOwned: conflicting})
	})
	assert.Equal(t, mapping, LookupError(errContextOwned))
}
