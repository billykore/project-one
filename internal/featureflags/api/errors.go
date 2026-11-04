package api

import (
	"net/http"

	"github.com/billykore/project-one/internal/featureflags/domain"
	pkgmiddleware "github.com/billykore/project-one/internal/pkg/api/middleware"
	"github.com/billykore/project-one/internal/pkg/problem"
)

// errorMappings maps the featureflags context's sentinel errors to problem details.
var errorMappings = map[error]pkgmiddleware.ErrorMapping{
	domain.ErrFlagNotFound:         {Status: http.StatusNotFound, Code: problem.CodeNotFound, TypeSlug: "not-found", Title: "Not Found", Detail: "Feature flag not found"},
	domain.ErrFlagKeyExists:        {Status: http.StatusConflict, Code: problem.CodeAlreadyExists, TypeSlug: "already-exists", Title: "Conflict", Detail: "Feature flag key already exists"},
	domain.ErrFlagArchived:         {Status: http.StatusConflict, Code: problem.CodeConflict, TypeSlug: "conflict", Title: "Conflict", Detail: "Feature flag is archived"},
	domain.ErrRevisionConflict:     {Status: http.StatusConflict, Code: problem.CodeConflict, TypeSlug: "conflict", Title: "Conflict", Detail: "Feature flag was changed; refresh and retry"},
	domain.ErrInvalidFlagMode:      {Status: http.StatusBadRequest, Code: problem.CodeInvalidArgument, TypeSlug: "invalid-argument", Title: "Bad Request", Detail: "Invalid feature flag mode"},
	domain.ErrFlagSettingNotFound:  {Status: http.StatusNotFound, Code: problem.CodeNotFound, TypeSlug: "not-found", Title: "Not Found", Detail: "Feature flag setting not found"},
	domain.ErrConflictingOverrides: {Status: http.StatusUnprocessableEntity, Code: problem.CodeInvalidArgument, TypeSlug: "invalid-argument", Title: "Unprocessable Entity", Detail: "User cannot be both included and excluded"},
	domain.ErrOperatorOnly:         {Status: http.StatusForbidden, Code: problem.CodePermissionDenied, TypeSlug: "permission-denied", Title: "Forbidden", Detail: "Operator access required"},
}
