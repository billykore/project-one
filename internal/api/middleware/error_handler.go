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

// ErrorHandler returns an echo.HTTPErrorHandler that maps domain errors to structured JSON responses.
// Set via e.HTTPErrorHandler = ErrorHandler(log, isProduction) in main.go.
//
// ponytail: no middleware wrapper needed. Echo passes errors directly to HTTPErrorHandler.
// Handler return nil → Echo never calls this. Return error → this formats and writes the response.
func ErrorHandler(log ports.Logger, isProduction bool) func(err error, c echo.Context) {
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

		// FR-008: if the error is already an *echo.HTTPError, use its status code.
		if httpErr, ok := errors.AsType[*echo.HTTPError](err); ok {
			status = httpErr.Code
		}

		// FR-009: validation errors always map to 400 regardless of sentinel matching.
		var validationErrs validator.ValidationErrors
		var hasValidationErrs bool
		if validationErrs, hasValidationErrs = errors.AsType[validator.ValidationErrors](err); hasValidationErrs {
			status = 400
			code = domain.CodeInvalidArgument
		}

		requestID := c.Response().Header().Get(echo.HeaderXRequestID)

		body := dto.APIErrorResponse{
			Error: dto.StructuredError{
				Code:      code,
				Message:   mapping.Message,
				RequestID: requestID,
				Details:   validationDetails(validationErrs),
			},
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

		if !isProduction {
			logFields = append(logFields, "stack_trace", string(debug.Stack()))
		}

		// Log the full error server-side. Client gets the sanitized message.
		logFields = append(logFields, "error", err.Error())

		if status >= 500 {
			log.Error(c.Request().Context(), "request error", logFields...)
		} else {
			log.Warn(c.Request().Context(), "request error", logFields...)
		}

		// FR-010: in production, never expose raw error. Message is already the registry-default.
		_ = c.JSON(status, body)
	}
}

// validationDetails converts validator.ValidationErrors to structured ErrorDetail objects.
func validationDetails(errs validator.ValidationErrors) []dto.ErrorDetail {
	details := make([]dto.ErrorDetail, 0, len(errs))
	for _, fe := range errs {
		details = append(details, dto.ErrorDetail{
			Field:   fe.Field(),
			Reason:  fe.Tag(),
			Message: tagMessage(fe.Field(), fe.Tag(), fe.Param()),
		})
	}
	return details
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
