package dto

import "time"

// LoginRequest is the request body for login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse is the response body for a successful login.
type LoginResponse struct {
	Message  string `json:"message"`
	Username string `json:"username"`
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

// ChangePasswordRequest is the request body for updating the password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=8"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// UpdateProfileRequest is the request body for updating user profile.
type UpdateProfileRequest struct {
	FirstName string `json:"first_name" validate:"required,min=3,max=100"`
	LastName  string `json:"last_name" validate:"required,min=3,max=100"`
	Username  string `json:"username" validate:"required,min=3,max=30"`
}

// UpdateProfileResponse is the response body for a successful profile update.
type UpdateProfileResponse struct {
	Message  string `json:"message"`
	Username string `json:"username"`
}

// SearchUsersRequest holds query parameters for the user search endpoint.
type SearchUsersRequest struct {
	Q      string `query:"q" validate:"required,min=3"`
	Cursor string `query:"cursor"`
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=20"`
}

// SearchUsersItem is a single search result item.
type SearchUsersItem struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

// SearchUsersResponse wraps the search results with cursor pagination.
type SearchUsersResponse struct {
	Data       []SearchUsersItem `json:"data"`
	NextCursor string            `json:"next_cursor"`
	HasMore    bool              `json:"has_more"`
}

// PostsListResponse wraps a cursor-paginated post list.
type PostsListResponse struct {
	Data       []PostResponse `json:"data"`
	NextCursor string         `json:"next_cursor"`
	HasMore    bool           `json:"has_more"`
}

// PostResponse is the compact publishing projection embedded in an identity
// profile response. It avoids coupling the identity HTTP contract to another
// context's DTO package.
type PostResponse struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
