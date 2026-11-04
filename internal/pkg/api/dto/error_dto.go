package dto

// ProblemDetail is the RFC 9457 (STD 97) Problem Details for HTTP APIs response body.
// Standard fields: type, title, status, detail, instance.
// Extension fields: code, request_id, errors.
type ProblemDetail struct {
	Type      string            `json:"type"`
	Title     string            `json:"title"`
	Status    int               `json:"status"`
	Detail    string            `json:"detail"`
	Instance  string            `json:"instance"`
	Code      string            `json:"code,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
	Errors    []ValidationError `json:"errors,omitempty"`
}

// ValidationError describes a single field-level validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}
