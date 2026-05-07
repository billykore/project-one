package handler

// ErrorResponse is the response body for a failed login.
type ErrorResponse struct {
	Error string `json:"error"`
}
