// Package middleware provides Echo middleware for authorization and error handling.
package middleware

import (
	"errors"
	"net/http"

	"github.com/billykore/project-one/internal/platform/problem"
	publishingdomain "github.com/billykore/project-one/internal/publishing/domain"
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

// ponytail: package-level map, no struct/constructor/Register. Register new errors here.
var errorMappings = map[error]ErrorMapping{
	echo.ErrNotFound:                  {http.StatusNotFound, problem.CodeNotFound, "not-found", "Not Found", "Resource not found"},
	echo.ErrUnauthorized:              {http.StatusUnauthorized, problem.CodeUnauthenticated, "unauthenticated", "Unauthorized", "Unauthorized"},
	echo.ErrForbidden:                 {http.StatusForbidden, problem.CodePermissionDenied, "permission-denied", "Forbidden", "Permission denied"},
	echo.ErrMethodNotAllowed:          {http.StatusMethodNotAllowed, problem.CodePermissionDenied, "method-not-allowed", "Method Not Allowed", "HTTP method not allowed"},
	echo.ErrInternalServerError:       {http.StatusInternalServerError, problem.CodeInternal, "internal-server-error", "Internal Server Error", "Internal server error"},
	echo.ErrBadRequest:                {http.StatusBadRequest, problem.CodeInvalidArgument, "invalid-argument", "Bad Request", "Invalid request"},
	problem.ErrUserNotFound:           {http.StatusNotFound, problem.CodeNotFound, "not-found", "Not Found", "User not found"},
	problem.ErrInvalidCredentials:     {http.StatusUnauthorized, problem.CodeUnauthenticated, "unauthenticated", "Unauthorized", "Invalid email or password"},
	problem.ErrEmailAlreadyRegistered: {http.StatusConflict, problem.CodeAlreadyExists, "already-exists", "Conflict", "Email is already registered"},
	problem.ErrAlreadyFollowing:       {http.StatusConflict, problem.CodeAlreadyExists, "already-exists", "Conflict", "Already following this user"},
	problem.ErrCannotFollowSelf:       {http.StatusUnprocessableEntity, problem.CodeInvalidArgument, "invalid-argument", "Unprocessable Entity", "Cannot follow yourself"},
	problem.ErrNotFollowing:           {http.StatusNotFound, problem.CodeNotFound, "not-found", "Not Found", "Not following this user"},
	problem.ErrCannotUnfollowSelf:     {http.StatusUnprocessableEntity, problem.CodeInvalidArgument, "invalid-argument", "Unprocessable Entity", "Cannot unfollow yourself"},
	problem.ErrUsernameAlreadyTaken:   {http.StatusConflict, problem.CodeAlreadyExists, "already-exists", "Conflict", "Username is already taken"},
	problem.ErrPostNotFound:           {http.StatusNotFound, problem.CodeNotFound, "not-found", "Not Found", "Post not found"},
	problem.ErrInvalidPost:            {http.StatusBadRequest, problem.CodeInvalidArgument, "invalid-argument", "Bad Request", "Invalid post data"},
	publishingdomain.ErrAlreadyLiked:  {http.StatusConflict, problem.CodeAlreadyExists, "already-exists", "Conflict", "Post already liked"},
	publishingdomain.ErrNotLiked:      {http.StatusNotFound, problem.CodeNotFound, "not-found", "Not Found", "Post not liked"},
	problem.ErrCommentNotFound:        {http.StatusNotFound, problem.CodeNotFound, "not-found", "Not Found", "Comment not found"},
	problem.ErrNotificationNotFound:   {http.StatusNotFound, problem.CodeNotFound, "not-found", "Not Found", "Notification not found"},
	problem.ErrInvalidNotification:    {http.StatusBadRequest, problem.CodeInvalidArgument, "invalid-argument", "Bad Request", "Invalid notification"},
	problem.ErrPostIDMustBeANumber:    {http.StatusBadRequest, problem.CodeInvalidArgument, "invalid-argument", "Bad Request", "Post ID must be a number"},
	problem.ErrSearchQueryTooShort:    {http.StatusBadRequest, problem.CodeInvalidArgument, "invalid-argument", "Bad Request", "Search query must be at least 3 characters"},
	problem.ErrInvalidCursor:          {http.StatusBadRequest, problem.CodeInvalidArgument, "invalid-argument", "Bad Request", "Invalid cursor"},
	problem.ErrFlagNotFound:           {http.StatusNotFound, problem.CodeNotFound, "not-found", "Not Found", "Feature flag not found"},
	problem.ErrFlagKeyExists:          {http.StatusConflict, problem.CodeAlreadyExists, "already-exists", "Conflict", "Feature flag key already exists"},
	problem.ErrFlagArchived:           {http.StatusConflict, problem.CodeConflict, "conflict", "Conflict", "Feature flag is archived"},
	problem.ErrRevisionConflict:       {http.StatusConflict, problem.CodeConflict, "conflict", "Conflict", "Feature flag was changed; refresh and retry"},
	problem.ErrInvalidFlagMode:        {http.StatusBadRequest, problem.CodeInvalidArgument, "invalid-argument", "Bad Request", "Invalid feature flag mode"},
	problem.ErrFlagSettingNotFound:    {http.StatusNotFound, problem.CodeNotFound, "not-found", "Not Found", "Feature flag setting not found"},
	problem.ErrConflictingOverrides:   {http.StatusUnprocessableEntity, problem.CodeInvalidArgument, "invalid-argument", "Unprocessable Entity", "User cannot be both included and excluded"},
	problem.ErrInvalidArgument:        {http.StatusUnprocessableEntity, problem.CodeInvalidArgument, "invalid-argument", "Unprocessable Entity", "Invalid argument"},
	problem.ErrOperatorOnly:           {http.StatusForbidden, problem.CodePermissionDenied, "permission-denied", "Forbidden", "Operator access required"},
	problem.ErrFeatureDisabled:        {http.StatusNotFound, problem.CodeNotFound, "not-found", "Not Found", "Feature not available"},
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
