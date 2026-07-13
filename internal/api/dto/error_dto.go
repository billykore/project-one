package dto

// StructuredError is the error object inside the structured error response.
type StructuredError struct {
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	RequestID string        `json:"request_id"`
	Details   []ErrorDetail `json:"details,omitempty"`
}

// APIErrorResponse is the new structured error response wrapper.
type APIErrorResponse struct {
	Error StructuredError `json:"error"`
}

// ErrorDetail describes a single field-level validation failure.
type ErrorDetail struct {
	Field   string `json:"field"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}
