package dto

// UserResponse is the response body for user data.
type UserResponse struct {
	ID    string `json:"user_id"`
	Email string `json:"email"`
}
