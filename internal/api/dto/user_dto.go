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

// LogoutResponse is the response body for a successful logout.
type LogoutResponse struct {
	Message string `json:"message"`
}

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

// UserResponse is the response body for user data.
type UserResponse struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// FollowResponse is the response body for a successful follow action.
type FollowResponse struct {
	Message string     `json:"message"`
	Data    FollowData `json:"data"`
}

// FollowData is the data part of the follow response.
type FollowData struct {
	FollowedUserID int    `json:"followed_user_id"`
	FollowedAt     string `json:"followed_at"`
}
