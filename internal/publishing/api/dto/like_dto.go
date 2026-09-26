package dto

// LikeResponse is the response body for like-related endpoints.
type LikeResponse struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"like_count"`
}
