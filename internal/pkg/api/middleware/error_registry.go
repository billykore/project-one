// Package middleware provides Echo middleware for authorization and error handling.
package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/billykore/project-one/internal/pkg/problem"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// ErrorMapping associates a sentinel error with its HTTP representation.
type ErrorMapping struct {
	// HTTP status code, e.g., 404.
	Status int
	// Application-specific error code, e.g., "NOT_FOUND".
	Code string
	// TypeSlug is the URI path segment after the base URL, e.g., "not-found". Empty → about:blank.
	TypeSlug string
	// Short human-readable summary, e.g., "Not Found".
	Title string
	// Human-readable detail message, e.g., "User not found".
	Detail string
}

// errorMappings starts with the transport and shared problem sentinels; each
// bounded context adds its own sentinels through RegisterErrorMappings.
var (
	errorMappingsMu sync.RWMutex
	errorMappings   = map[error]ErrorMapping{
		echo.ErrNotFound:            {http.StatusNotFound, problem.CodeNotFound, "not-found", "Not Found", "Resource not found"},
		echo.ErrUnauthorized:        {http.StatusUnauthorized, problem.CodeUnauthenticated, "unauthenticated", "Unauthorized", "Unauthorized"},
		echo.ErrForbidden:           {http.StatusForbidden, problem.CodePermissionDenied, "permission-denied", "Forbidden", "Permission denied"},
		echo.ErrMethodNotAllowed:    {http.StatusMethodNotAllowed, problem.CodePermissionDenied, "method-not-allowed", "Method Not Allowed", "HTTP method not allowed"},
		echo.ErrInternalServerError: {http.StatusInternalServerError, problem.CodeInternal, "internal-server-error", "Internal Server Error", "Internal server error"},
		echo.ErrBadRequest:          {http.StatusBadRequest, problem.CodeInvalidArgument, "invalid-argument", "Bad Request", "Invalid request"},
		problem.ErrUserNotFound:     {http.StatusNotFound, problem.CodeNotFound, "not-found", "Not Found", "User not found"},
		problem.ErrInvalidCursor:    {http.StatusBadRequest, problem.CodeInvalidArgument, "invalid-argument", "Bad Request", "Invalid cursor"},
		problem.ErrInvalidArgument:  {http.StatusUnprocessableEntity, problem.CodeInvalidArgument, "invalid-argument", "Unprocessable Entity", "Invalid argument"},
		problem.ErrFeatureDisabled:  {http.StatusNotFound, problem.CodeNotFound, "not-found", "Not Found", "Feature not available"},
	}
)

// RegisterErrorMappings adds a context's sentinel errors to the shared lookup.
// Contexts call it while registering their routes, before the server starts.
// Registering the same mapping again is a no-op; a conflicting mapping for an
// already registered sentinel panics because it is a programming error.
func RegisterErrorMappings(mappings map[error]ErrorMapping) {
	errorMappingsMu.Lock()
	defer errorMappingsMu.Unlock()
	for sentinel, mapping := range mappings {
		if existing, ok := errorMappings[sentinel]; ok && existing != mapping {
			panic(fmt.Sprintf("conflicting error mapping for %q", sentinel))
		}
		errorMappings[sentinel] = mapping
	}
}

var defaultMapping = ErrorMapping{
	Status:   http.StatusInternalServerError,
	Code:     problem.CodeInternal,
	TypeSlug: "",
	Title:    "Internal Server Error",
	Detail:   "Something went wrong",
}

// LookupError walks the error chain with errors.Is and returns the matching ErrorMapping.
// Returns the default 500 mapping if no sentinel is recognized.
func LookupError(err error) ErrorMapping {
	errorMappingsMu.RLock()
	defer errorMappingsMu.RUnlock()
	for sentinel, m := range errorMappings {
		// ponytail: unordered map iteration; first match wins. With disjoint sentinels this is deterministic.
		if errors.Is(err, sentinel) {
			return m
		}
	}
	return defaultMapping
}

// StatusForError resolves the HTTP status the ErrorHandler writes for err.
// The metrics middleware shares this mapping so the observed status class
// always matches the response the client actually receives.
func StatusForError(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if httpErr, ok := errors.AsType[*echo.HTTPError](err); ok {
		return httpErr.Code
	}
	if _, isValidation := errors.AsType[validator.ValidationErrors](err); isValidation {
		return http.StatusBadRequest
	}
	return LookupError(err).Status
}
