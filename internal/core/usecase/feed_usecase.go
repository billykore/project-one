package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	"github.com/billykore/project-one/internal/pkg/pagination"
)

type feedUseCase struct {
	postRepo   ports.PostRepository
	followRepo ports.FollowRepository
	userRepo   ports.UserRepository
	log        ports.Logger
}

// NewFeedUseCase creates a new instance of FeedUseCase.
func NewFeedUseCase(
	postRepo ports.PostRepository,
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

func (u *feedUseCase) GetFeed(ctx context.Context, username string, cursor *pagination.Cursor, limit int) (*ports.FeedResult, error) {
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

	// Get followed usernames.
	followedUsernames, err := u.followRepo.GetFollowedUsernames(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get followed usernames: %w", err)
	}

	// Build username list: self + followed.
	usernames := append([]string{username}, followedUsernames...)

	// Extract cursor params.
	var cursorCreatedAt time.Time
	var cursorID int
	if cursor != nil {
		cursorCreatedAt = cursor.CreatedAt
		cursorID = cursor.ID
	}

	// Fetch one extra to detect has_more.
	dbLimit := limit + 1
	posts, err := u.postRepo.GetFeed(ctx, usernames, cursorCreatedAt, cursorID, dbLimit)
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

	// Build next cursor from last post.
	if len(result.Posts) > 0 {
		last := result.Posts[len(result.Posts)-1]
		result.NextCursor = &pagination.Cursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		}
	}

	return result, nil
}
