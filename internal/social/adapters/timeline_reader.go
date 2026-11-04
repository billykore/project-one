package repository

import (
	"context"

	vo "github.com/billykore/project-one/internal/pkg/pagination"
	publishingdomain "github.com/billykore/project-one/internal/publishing/domain"
	"github.com/billykore/project-one/internal/social/domain"
	"github.com/billykore/project-one/internal/social/ports"
)

// publishingPosts is the slice of publishing's post query repository this adapter reads.
type publishingPosts interface {
	GetFeed(ctx context.Context, userIDs []int, cursor *vo.Cursor, limit int) ([]*publishingdomain.Post, error)
}

// timelineReader translates publishing posts into social's own TimelinePost model
// so the publishing aggregate does not leak past this adapter.
type timelineReader struct {
	posts publishingPosts
}

var _ ports.TimelineReader = (*timelineReader)(nil)

// NewTimelineReader adapts publishing's post query repository to the TimelineReader port.
func NewTimelineReader(posts publishingPosts) ports.TimelineReader {
	return &timelineReader{posts: posts}
}

// GetFeed returns the social projection of the publishing posts.
func (r *timelineReader) GetFeed(ctx context.Context, userIDs []int, cursor *vo.Cursor, limit int) ([]*domain.TimelinePost, error) {
	posts, err := r.posts.GetFeed(ctx, userIDs, cursor, limit)
	if err != nil {
		return nil, err
	}
	timeline := make([]*domain.TimelinePost, 0, len(posts))
	for _, post := range posts {
		timeline = append(timeline, &domain.TimelinePost{
			ID:        post.ID,
			Author:    post.Username,
			Title:     post.Title,
			Content:   post.Content,
			Tags:      post.Tags,
			LikeCount: post.LikeCount,
			CreatedAt: post.CreatedAt,
			UpdatedAt: post.UpdatedAt,
		})
	}
	return timeline, nil
}
