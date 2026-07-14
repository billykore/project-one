// Package middleware provides Echo middleware for authorization and error handling.
package middleware

import (
	"errors"
	"net/http"

	"github.com/billykore/project-one/internal/core/domain"
)

// ErrorMapping associates a sentinel error with its HTTP representation.
type ErrorMapping struct {
	Status   int
	Code     string
	TypeSlug string // URI path segment after base URL, e.g., "not-found". Empty → about:blank.
	Title    string // Short human-readable summary, e.g., "Not Found".
	Detail   string // Human-readable detail message, e.g., "User not found".
}

// ponytail: package-level map, no struct/constructor/Register. Register new errors here.
var errorMappings = map[error]ErrorMapping{
	domain.ErrUserNotFound:           {http.StatusNotFound, domain.CodeNotFound, "not-found", "Not Found", "User not found"},
	domain.ErrInvalidCredentials:     {http.StatusUnauthorized, domain.CodeUnauthenticated, "unauthenticated", "Unauthorized", "Invalid email or password"},
	domain.ErrUnauthorized:           {http.StatusUnauthorized, domain.CodeUnauthenticated, "unauthenticated", "Unauthorized", "Unauthorized"},
	domain.ErrInternalServer:         {http.StatusInternalServerError, domain.CodeInternal, "", "Internal Server Error", "Internal server error"},
	domain.ErrEmailAlreadyRegistered: {http.StatusConflict, domain.CodeAlreadyExists, "already-exists", "Conflict", "Email is already registered"},
	domain.ErrValidationFailed:       {http.StatusBadRequest, domain.CodeInvalidArgument, "invalid-argument", "Bad Request", "Validation failed"},
	domain.ErrAlreadyFollowing:       {http.StatusConflict, domain.CodeAlreadyExists, "already-exists", "Conflict", "Already following this user"},
	domain.ErrCannotFollowSelf:       {http.StatusUnprocessableEntity, domain.CodeInvalidArgument, "invalid-argument", "Unprocessable Entity", "Cannot follow yourself"},
	domain.ErrNotFollowing:           {http.StatusNotFound, domain.CodeNotFound, "not-found", "Not Found", "Not following this user"},
	domain.ErrCannotUnfollowSelf:     {http.StatusUnprocessableEntity, domain.CodeInvalidArgument, "invalid-argument", "Unprocessable Entity", "Cannot unfollow yourself"},
	domain.ErrUsernameAlreadyTaken:   {http.StatusConflict, domain.CodeAlreadyExists, "already-exists", "Conflict", "Username is already taken"},
	domain.ErrPostNotFound:           {http.StatusNotFound, domain.CodeNotFound, "not-found", "Not Found", "Post not found"},
	domain.ErrInvalidPost:            {http.StatusBadRequest, domain.CodeInvalidArgument, "invalid-argument", "Bad Request", "Invalid post data"},
	domain.ErrAlreadyLiked:           {http.StatusConflict, domain.CodeAlreadyExists, "already-exists", "Conflict", "Post already liked"},
	domain.ErrNotLiked:               {http.StatusNotFound, domain.CodeNotFound, "not-found", "Not Found", "Post not liked"},
	domain.ErrCommentNotFound:        {http.StatusNotFound, domain.CodeNotFound, "not-found", "Not Found", "Comment not found"},
	domain.ErrNotificationNotFound:   {http.StatusNotFound, domain.CodeNotFound, "not-found", "Not Found", "Notification not found"},
	domain.ErrInvalidNotification:    {http.StatusBadRequest, domain.CodeInvalidArgument, "invalid-argument", "Bad Request", "Invalid notification"},
}

var defaultMapping = ErrorMapping{
	Status:   http.StatusInternalServerError,
	Code:     domain.CodeInternal,
	TypeSlug: "",
	Title:    "Internal Server Error",
	Detail:   "Internal server error",
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
