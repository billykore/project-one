package dto

// LoginRequest is the request body for login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse is the response body for a successful login.
type LoginResponse struct {
	Message string `json:"message"`
}

// ErrorResponse is the response body for a failed login.
type ErrorResponse struct {
	Error string `json:"error"`
}
