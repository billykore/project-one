// Package middleware provides Echo middleware for authorization and error handling.
package middleware

import (
	"errors"
	"runtime/debug"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// ErrorHandler returns an echo.HTTPErrorHandler that maps domain errors to RFC 9457 Problem Details responses.
// Set via e.HTTPErrorHandler = ErrorHandler(log, errorTypeBaseURL, isProduction) in main.go.
//
// ponytail: no middleware wrapper needed. Echo passes errors directly to HTTPErrorHandler.
// Handler return nil → Echo never calls this. Return error → this formats and writes the response.
func ErrorHandler(log ports.Logger, errorTypeBaseURL string, withStackTrace bool) func(err error, c echo.Context) {
	// ponytail: normalize base URL once at construction time.
	if errorTypeBaseURL == "" {
		errorTypeBaseURL = "http://localhost:8080/errors"
	}
	// Ensure trailing slash for clean concatenation.
	if errorTypeBaseURL[len(errorTypeBaseURL)-1] != '/' {
		errorTypeBaseURL += "/"
	}

	return func(err error, c echo.Context) {
		if c.Response().Committed {
			// Response already written — handler bug. Log and do nothing.
			log.Warn(c.Request().Context(), "error after response committed, cannot write error body",
				"error", err.Error(),
				"path", c.Request().URL.Path,
			)
			return
		}

		mapping := LookupError(err)
		status := mapping.Status
		code := mapping.Code

		// if the error is already an *echo.HTTPError, use its status code.
		if httpErr, ok := errors.AsType[*echo.HTTPError](err); ok {
			status = httpErr.Code
		}

		// validation errors always map to 400 regardless of sentinel matching.
		var validationErrs validator.ValidationErrors
		var hasValidationErrs bool
		if validationErrs, hasValidationErrs = errors.AsType[validator.ValidationErrors](err); hasValidationErrs {
			status = 400
			code = domain.CodeInvalidArgument
			mapping.Title = "Bad Request"
			mapping.Detail = "Invalid request"
		}

		requestID := c.Response().Header().Get(echo.HeaderXRequestID)

		// Construct RFC 9457 type URI from configurable base + mapping slug.
		typeURI := "about:blank"
		if mapping.TypeSlug != "" {
			typeURI = errorTypeBaseURL + mapping.TypeSlug
		}

		body := dto.ProblemDetail{
			Type:      typeURI,
			Title:     mapping.Title,
			Status:    status,
			Detail:    mapping.Detail,
			Instance:  c.Request().URL.Path,
			Code:      code,
			RequestID: requestID,
			Errors:    validationErrors(validationErrs),
		}

		// Log the error with structured fields.
		logFields := []any{
			"request_id", requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.Path,
			"status", status,
			"error_code", code,
		}

		// ponytail: extract username for log; never expose in error response.
		if username, ok := c.Get("username").(string); ok && username != "" {
			logFields = append(logFields, "user", username)
		} else {
			logFields = append(logFields, "user", "anonymous")
		}

		if withStackTrace {
			logFields = append(logFields, "stack_trace", string(debug.Stack()))
		}

		// Log the full error server-side. Client gets the sanitized message.
		logFields = append(logFields, "error", err.Error())

		if status >= 500 {
			log.Error(c.Request().Context(), "request error", logFields...)
		} else {
			log.Warn(c.Request().Context(), "request error", logFields...)
		}

		// Set RFC 9457 Content-Type before writing body.
		c.Response().Header().Set(echo.HeaderContentType, "application/problem+json")
		_ = c.JSON(status, body)
	}
}

// validationErrors converts validator.ValidationErrors to RFC 9457 extension items.
func validationErrors(errs validator.ValidationErrors) []dto.ValidationError {
	items := make([]dto.ValidationError, 0, len(errs))
	for _, fe := range errs {
		items = append(items, dto.ValidationError{
			Field:   fe.Field(),
			Reason:  fe.Tag(),
			Message: tagMessage(fe.Field(), fe.Tag(), fe.Param()),
		})
	}
	return items
}

// ponytail: simple lookup table for common validator tags instead of pulling in translator dependency.
func tagMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return field + " is required"
	case "min":
		return field + " must be at least " + param + " characters"
	case "max":
		return field + " must be at most " + param + " characters"
	case "email":
		return field + " must be a valid email address"
	default:
		return field + " failed validation: " + tag
	}
}
