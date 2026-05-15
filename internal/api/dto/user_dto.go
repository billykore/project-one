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
	Username  string `json:"username" validate:"required,min=3"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

// RegisterResponse is the response body for a successful registration.
type RegisterResponse struct {
	Message string `json:"message"`
}

// UserResponse is the response body for user data.
type UserResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

// GetFollowingRequest is the query parameters for getting following list.
type GetFollowingRequest struct {
	Limit  int `query:"limit" validate:"omitempty,min=1,max=100"`
	Offset int `query:"offset" validate:"omitempty,min=0"`
}

// FollowingResponse is the response body for a user being followed.
type FollowingResponse struct {
	Username   string `json:"username"`
	Name       string `json:"name"`
	FollowedAt string `json:"followed_at"`
	IsMutual   bool   `json:"is_mutual"`
}

// GetFollowersRequest is the query parameters for getting followers list.
type GetFollowersRequest struct {
	Limit  int `query:"limit" validate:"omitempty,min=1,max=100"`
	Offset int `query:"offset" validate:"omitempty,min=0"`
}

// FollowerResponse is the response body for a user following.
type FollowerResponse struct {
	Username   string `json:"username"`
	Name       string `json:"name"`
	FollowedAt string `json:"followed_at"`
	IsMutual   bool   `json:"is_mutual"`
}

// UnfollowResponse is the response body for a successful unfollow action.
type UnfollowResponse struct {
	Message string `json:"message"`
}

// FollowResponse is the response body for a successful follow action.
type FollowResponse struct {
	Message string     `json:"message"`
	Data    FollowData `json:"data"`
}

// FollowData is the data part of the follow response.
type FollowData struct {
	FollowedUsername string `json:"followed_username"`
	FollowedAt       string `json:"followed_at"`
}
