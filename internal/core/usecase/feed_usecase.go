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
	userRepo   ports.UserRepository
	log        ports.Logger
}

// NewFeedUseCase creates a new instance of FeedUseCase.
func NewFeedUseCase(
	postRepo ports.PostQueryRepository,
	followRepo ports.FollowRepository,
	userRepo ports.UserRepository,
	log ports.Logger,
) ports.FeedUseCase {
	if postRepo == nil || followRepo == nil || userRepo == nil || log == nil {
		panic("NewFeedUseCase: dependencies must not be nil")
	}
	return &feedUseCase{
		postRepo:   postRepo,
		followRepo: followRepo,
		userRepo:   userRepo,
		log:        log,
	}
}

func (u *feedUseCase) GetFeed(ctx context.Context, username string, cursor *vo.Cursor, limit int) (*ports.FeedResult, error) {
	// Clamp limit.
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Resolve user.
	user, err := u.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Resolve the social graph by stable identity, not mutable usernames.
	followedUserIDs, err := u.followRepo.GetFollowedUserIDs(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get followed user IDs: %w", err)
	}

	// Build author ID list: self + followed.
	userIDs := append([]int{user.ID}, followedUserIDs...)

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
