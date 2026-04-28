package dto

// RegisterRequest is the request body for registration.
type RegisterRequest struct {
	FirstName string `json:"first_name" validate:"required,min=3"`
	LastName  string `json:"last_name" validate:"required,min=3"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

// RegisterResponse is the response body for a successful registration.
type RegisterResponse struct {
	Message string `json:"message"`
}
