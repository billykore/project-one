// Package domain defines core business entities and sentinel error codes.
package domain

// Error codes are general, HTTP-level machine-readable identifiers for API error responses.
// Stable across releases — safe for programmatic dispatch by API consumers.
const (
	CodeNotFound         = "NOT_FOUND"
	CodeInvalidArgument  = "INVALID_ARGUMENT"
	CodeUnauthenticated  = "UNAUTHENTICATED"
	CodePermissionDenied = "PERMISSION_DENIED"
	CodeAlreadyExists    = "ALREADY_EXISTS"
	CodeInternal         = "INTERNAL"
)
