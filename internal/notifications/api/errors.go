package api

import (
	"net/http"

	"github.com/billykore/project-one/internal/notifications/domain"
	pkgmiddleware "github.com/billykore/project-one/internal/pkg/api/middleware"
	"github.com/billykore/project-one/internal/pkg/problem"
)

// errorMappings maps the notifications context's sentinel errors to problem details.
var errorMappings = map[error]pkgmiddleware.ErrorMapping{
	domain.ErrNotificationNotFound: {Status: http.StatusNotFound, Code: problem.CodeNotFound, TypeSlug: "not-found", Title: "Not Found", Detail: "Notification not found"},
	domain.ErrInvalidNotification:  {Status: http.StatusBadRequest, Code: problem.CodeInvalidArgument, TypeSlug: "invalid-argument", Title: "Bad Request", Detail: "Invalid notification"},
}
