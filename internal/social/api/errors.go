package api

import (
	"net/http"

	pkgmiddleware "github.com/billykore/project-one/internal/pkg/api/middleware"
	"github.com/billykore/project-one/internal/pkg/problem"
	"github.com/billykore/project-one/internal/social/domain"
)

// errorMappings maps the social context's sentinel errors to problem details.
var errorMappings = map[error]pkgmiddleware.ErrorMapping{
	domain.ErrAlreadyFollowing:   {Status: http.StatusConflict, Code: problem.CodeAlreadyExists, TypeSlug: "already-exists", Title: "Conflict", Detail: "Already following this user"},
	domain.ErrCannotFollowSelf:   {Status: http.StatusUnprocessableEntity, Code: problem.CodeInvalidArgument, TypeSlug: "invalid-argument", Title: "Unprocessable Entity", Detail: "Cannot follow yourself"},
	domain.ErrNotFollowing:       {Status: http.StatusNotFound, Code: problem.CodeNotFound, TypeSlug: "not-found", Title: "Not Found", Detail: "Not following this user"},
	domain.ErrCannotUnfollowSelf: {Status: http.StatusUnprocessableEntity, Code: problem.CodeInvalidArgument, TypeSlug: "invalid-argument", Title: "Unprocessable Entity", Detail: "Cannot unfollow yourself"},
}
