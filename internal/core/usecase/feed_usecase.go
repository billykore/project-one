package usecase

import (
	"context"
	"fmt"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	vo "github.com/billykore/project-one/internal/core/valueobject"
)

type feedUseCase struct {
	postRepo   ports.PostQueryRepository
	followRepo ports.FollowRepository
	log        ports.Logger
}

// NewFeedUseCase creates a new instance of FeedUseCase.
func NewFeedUseCase(
	postRepo ports.PostQueryRepository,
	followRepo ports.FollowRepository,
	log ports.Logger,
) ports.FeedUseCase {
	if postRepo == nil || followRepo == nil || log == nil {
		panic("NewFeedUseCase: dependencies must not be nil")
	}
	return &feedUseCase{
		postRepo:   postRepo,
		followRepo: followRepo,
		log:        log,
	}
}

func (u *feedUseCase) GetFeed(ctx context.Context, userID int, cursor *vo.Cursor, limit int) (*ports.FeedResult, error) {
	if userID <= 0 {
		return nil, domain.ErrInvalidUser
	}
	// Clamp limit.
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Resolve the social graph by stable identity, not mutable usernames.
	followedUserIDs, err := u.followRepo.GetFollowedUserIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get followed user IDs: %w", err)
	}

	// Build author ID list: self + followed.
	userIDs := append([]int{userID}, followedUserIDs...)

	// Fetch one extra to detect has_more.
	dbLimit := limit + 1
	posts, err := u.postRepo.GetFeed(ctx, userIDs, cursor, dbLimit)
	if err != nil {
		return nil, fmt.Errorf("get feed from repo: %w", err)
	}

	result := &ports.FeedResult{
		Posts:   posts,
		HasMore: false,
	}

	if len(posts) == dbLimit {
		result.HasMore = true
		result.Posts = posts[:limit]
	}

	// Only return a cursor when another page exists.
	if result.HasMore && len(result.Posts) > 0 {
		last := result.Posts[len(result.Posts)-1]
		result.NextCursor = &vo.Cursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		}
	}

	return result, nil
}
