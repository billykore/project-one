package repository

import (
	"context"
	"fmt"

	"github.com/billykore/project-one/internal/identity/domain"
	"github.com/billykore/project-one/internal/identity/ports"
	vo "github.com/billykore/project-one/internal/platform/pagination"
	"github.com/billykore/project-one/internal/platform/problem"
	"gorm.io/gorm"
)

type userSearchRepository struct {
	db *gorm.DB
}

// NewUserSearchRepository creates a new instance of UserSearchRepository.
func NewUserSearchRepository(db *gorm.DB) ports.UserSearchRepository {
	return &userSearchRepository{db: db}
}

// Search executes a hybrid query that ranks results:
// 1. exact match (username = query)
// 2. prefix match (username LIKE 'query%')
// 3. trigram fuzzy match (pg_trgm similarity)
// Within each tier, results are ordered by trigram similarity descending and
// username ascending. The cursor stores the final result's sort values.
func (r *userSearchRepository) Search(ctx context.Context, query string, cursor *vo.Cursor, limit int) ([]domain.SearchResult, *vo.Cursor, bool, error) {
	type row struct {
		Username        string
		FirstName       string
		LastName        string
		MatchRank       int
		SimilarityScore float64
	}

	// Fetch limit+1 to detect has_more.
	var rows []row
	statement := `
WITH ranked AS (
	SELECT
		username,
		first_name,
		last_name,
		CASE WHEN username = ? THEN 0
		     WHEN username LIKE ? || '%' THEN 1
		     ELSE 2
		END AS match_rank,
		similarity(username, ?) AS similarity_score
	FROM users
	WHERE username % ?
	   OR username LIKE ? || '%'
)
SELECT username, first_name, last_name, match_rank, similarity_score
FROM ranked`
	args := []any{query, query, query, query, query}
	if cursor != nil {
		statement += `
WHERE match_rank > ?
   OR (match_rank = ? AND similarity_score < ?)
   OR (match_rank = ? AND similarity_score = ? AND username > ?)`
		args = append(args, cursor.Rank, cursor.Rank, cursor.Score, cursor.Rank, cursor.Score, cursor.Key)
	}
	statement += `
ORDER BY match_rank ASC, similarity_score DESC, username ASC
LIMIT ?`
	args = append(args, limit+1)
	err := r.db.WithContext(ctx).Raw(statement, args...).Scan(&rows).Error
	if err != nil {
		return nil, nil, false, fmt.Errorf("%w: %v", problem.ErrRepositoryFailure, err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	results := make([]domain.SearchResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, domain.SearchResult{
			Username:  r.Username,
			FirstName: r.FirstName,
			LastName:  r.LastName,
			Rank:      r.MatchRank,
			Score:     r.SimilarityScore,
		})
	}

	var nextCursor *vo.Cursor
	if hasMore {
		last := results[len(results)-1]
		nextCursor = &vo.Cursor{Key: last.Username, Rank: last.Rank, Score: last.Score}
	}

	return results, nextCursor, hasMore, nil
}
