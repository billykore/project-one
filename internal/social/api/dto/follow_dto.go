package dto

// GetFollowingRequest is the query parameters for getting a following list.
type GetFollowingRequest struct {
	Cursor string `query:"cursor"`
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=100"`
}

// FollowingResponse is the response body for a user being followed.
type FollowingResponse struct {
	Username   string `json:"username"`
	Name       string `json:"name"`
	FollowedAt string `json:"followed_at"`
	IsMutual   bool   `json:"is_mutual"`
}

// FollowingListResponse wraps a cursor-paginated following list.
type FollowingListResponse struct {
	Data       []FollowingResponse `json:"data"`
	NextCursor string              `json:"next_cursor"`
	HasMore    bool                `json:"has_more"`
}

// GetFollowersRequest is the query parameters for getting a followers list.
type GetFollowersRequest struct {
	Cursor string `query:"cursor"`
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=100"`
}

// FollowerResponse is the response body for a user following.
type FollowerResponse struct {
	Username   string `json:"username"`
	Name       string `json:"name"`
	FollowedAt string `json:"followed_at"`
	IsMutual   bool   `json:"is_mutual"`
}

// FollowersListResponse wraps a cursor-paginated followers list.
type FollowersListResponse struct {
	Data       []FollowerResponse `json:"data"`
	NextCursor string             `json:"next_cursor"`
	HasMore    bool               `json:"has_more"`
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
