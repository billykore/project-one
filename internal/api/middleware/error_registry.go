// Package middleware provides Echo middleware for authorization and error handling.
package middleware

import (
	"errors"
	"net/http"

	"github.com/billykore/project-one/internal/core/domain"
)

// ErrorMapping associates a sentinel error with its HTTP representation.
type ErrorMapping struct {
	Status  int
	Code    string
	Message string
}

// ponytail: package-level map, no struct/constructor/Register. Register new errors here.
var errorMappings = map[error]ErrorMapping{
	domain.ErrUserNotFound:           {http.StatusNotFound, domain.CodeNotFound, "User not found"},
	domain.ErrInvalidCredentials:     {http.StatusUnauthorized, domain.CodeUnauthenticated, "Invalid email or password"},
	domain.ErrUnauthorized:           {http.StatusUnauthorized, domain.CodeUnauthenticated, "Unauthorized"},
	domain.ErrInternalServer:         {http.StatusInternalServerError, domain.CodeInternal, "Internal server error"},
	domain.ErrEmailAlreadyRegistered: {http.StatusConflict, domain.CodeAlreadyExists, "Email is already registered"},
	domain.ErrValidationFailed:       {http.StatusBadRequest, domain.CodeInvalidArgument, "Validation failed"},
	domain.ErrAlreadyFollowing:       {http.StatusConflict, domain.CodeAlreadyExists, "Already following this user"},
	domain.ErrCannotFollowSelf:       {http.StatusUnprocessableEntity, domain.CodeInvalidArgument, "Cannot follow yourself"},
	domain.ErrNotFollowing:           {http.StatusNotFound, domain.CodeNotFound, "Not following this user"},
	domain.ErrCannotUnfollowSelf:     {http.StatusUnprocessableEntity, domain.CodeInvalidArgument, "Cannot unfollow yourself"},
	domain.ErrUsernameAlreadyTaken:   {http.StatusConflict, domain.CodeAlreadyExists, "Username is already taken"},
	domain.ErrPostNotFound:           {http.StatusNotFound, domain.CodeNotFound, "Post not found"},
	domain.ErrInvalidPost:            {http.StatusBadRequest, domain.CodeInvalidArgument, "Invalid post data"},
	domain.ErrAlreadyLiked:           {http.StatusConflict, domain.CodeAlreadyExists, "Post already liked"},
	domain.ErrNotLiked:               {http.StatusNotFound, domain.CodeNotFound, "Post not liked"},
	domain.ErrCommentNotFound:        {http.StatusNotFound, domain.CodeNotFound, "Comment not found"},
	domain.ErrNotificationNotFound:   {http.StatusNotFound, domain.CodeNotFound, "Notification not found"},
	domain.ErrInvalidNotification:    {http.StatusBadRequest, domain.CodeInvalidArgument, "Invalid notification"},
}

var defaultMapping = ErrorMapping{
	Status:  http.StatusInternalServerError,
	Code:    domain.CodeInternal,
	Message: "Internal server error",
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
