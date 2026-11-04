package api

import (
	"net/http"

	"github.com/billykore/project-one/internal/identity/domain"
	pkgmiddleware "github.com/billykore/project-one/internal/pkg/api/middleware"
	"github.com/billykore/project-one/internal/pkg/problem"
)

// errorMappings maps the identity context's sentinel errors to problem details.
var errorMappings = map[error]pkgmiddleware.ErrorMapping{
	domain.ErrInvalidCredentials:     {Status: http.StatusUnauthorized, Code: problem.CodeUnauthenticated, TypeSlug: "unauthenticated", Title: "Unauthorized", Detail: "Invalid email or password"},
	domain.ErrEmailAlreadyRegistered: {Status: http.StatusConflict, Code: problem.CodeAlreadyExists, TypeSlug: "already-exists", Title: "Conflict", Detail: "Email is already registered"},
	domain.ErrUsernameAlreadyTaken:   {Status: http.StatusConflict, Code: problem.CodeAlreadyExists, TypeSlug: "already-exists", Title: "Conflict", Detail: "Username is already taken"},
	domain.ErrSearchQueryTooShort:    {Status: http.StatusBadRequest, Code: problem.CodeInvalidArgument, TypeSlug: "invalid-argument", Title: "Bad Request", Detail: "Search query must be at least 3 characters"},
}
