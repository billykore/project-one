package dto

import (
	"time"

	"github.com/billykore/project-one/internal/core/ports"
)

// FeedResponse is the envelope for the GET /feeds response.
type FeedResponse struct {
	Data       []PostItem `json:"data"`
	NextCursor string     `json:"next_cursor"`
	HasMore    bool       `json:"has_more"`
}

// PostItem is a feed post without nested comments.
type PostItem struct {
	ID        int       `json:"id"`
	Author    string    `json:"author"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	LikeCount int       `json:"like_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToFeedResponse converts a FeedResult to the API response DTO.
func ToFeedResponse(result *ports.FeedResult) FeedResponse {
	data := make([]PostItem, 0, len(result.Posts))
	for _, p := range result.Posts {
		data = append(data, PostItem{
			ID:        p.ID,
			Author:    p.Username,
			Title:     p.Title,
			Content:   p.Content,
			Tags:      p.Tags,
			LikeCount: p.LikeCount,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		})
	}

	var nextCursor string
	if result.NextCursor != nil {
		nextCursor = result.NextCursor.Encode()
	}

	return FeedResponse{
		Data:       data,
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	}
}
