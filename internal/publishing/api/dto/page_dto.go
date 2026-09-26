package dto

// PostsListResponse is the paginated publishing response used by post queries.
type PostsListResponse struct {
	Data       []PostResponse `json:"data"`
	NextCursor string         `json:"next_cursor"`
	HasMore    bool           `json:"has_more"`
}
