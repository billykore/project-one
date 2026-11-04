package api

import (
	"net/http"

	pkgmiddleware "github.com/billykore/project-one/internal/pkg/api/middleware"
	"github.com/billykore/project-one/internal/pkg/problem"
	"github.com/billykore/project-one/internal/publishing/domain"
)

// errorMappings maps the publishing context's sentinel errors to problem details.
var errorMappings = map[error]pkgmiddleware.ErrorMapping{
	domain.ErrPostNotFound:        {Status: http.StatusNotFound, Code: problem.CodeNotFound, TypeSlug: "not-found", Title: "Not Found", Detail: "Post not found"},
	domain.ErrInvalidPost:         {Status: http.StatusBadRequest, Code: problem.CodeInvalidArgument, TypeSlug: "invalid-argument", Title: "Bad Request", Detail: "Invalid post data"},
	domain.ErrAlreadyLiked:        {Status: http.StatusConflict, Code: problem.CodeAlreadyExists, TypeSlug: "already-exists", Title: "Conflict", Detail: "Post already liked"},
	domain.ErrNotLiked:            {Status: http.StatusNotFound, Code: problem.CodeNotFound, TypeSlug: "not-found", Title: "Not Found", Detail: "Post not liked"},
	domain.ErrCommentNotFound:     {Status: http.StatusNotFound, Code: problem.CodeNotFound, TypeSlug: "not-found", Title: "Not Found", Detail: "Comment not found"},
	domain.ErrPostIDMustBeANumber: {Status: http.StatusBadRequest, Code: problem.CodeInvalidArgument, TypeSlug: "invalid-argument", Title: "Bad Request", Detail: "Post ID must be a number"},
}
