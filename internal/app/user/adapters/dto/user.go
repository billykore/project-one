package dto

// UserResponse is the response body for user data.
type UserResponse struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}
