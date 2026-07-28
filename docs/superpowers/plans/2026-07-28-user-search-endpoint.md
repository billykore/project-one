# User Search Endpoint — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `GET /users/search?q=<query>&limit=<n>` — a public endpoint that searches users by username using PostgreSQL trigram and prefix indexes.

**Architecture:** Follows existing Clean Architecture layers. New `UserSearchRepository` driven port (separate from `UserRepository`), `SearchUsers` use case method, hybrid SQL query (exact → prefix → trigram), minimal DTO (`username` + `name`). No auth required.

**Tech Stack:** Go 1.26+, Echo, GORM, PostgreSQL with pg_trgm extension, GoMock, Testify

## Global Constraints

- Public endpoint (no auth middleware)
- Query `q`: required, minimum 3 characters
- Query `limit`: optional, 1–20, default 10
- Response: `{ "data": [{ "username": "...", "name": "..." }] }`
- Empty results return 200 with empty array (not 404)
- All errors follow RFC 9457 Problem Details format
- Follow existing TDD patterns: test first, then implementation
- GoMock generated mocks via `make mocks`
- Swagger docs via `make docs`

---

### Task 1: Add Domain Types (SearchResult VO and sentinel error)

**Files:**
- Modify: `internal/core/domain/user.go:19-19` (append after `User` struct)
- Modify: `internal/core/domain/errors.go:52-52` (append before last `)`)

**Interfaces:**
- Produces: `domain.SearchResult` struct with `Name()` method, `domain.ErrSearchQueryTooShort` sentinel error

- [ ] **Step 1: Add SearchResult value object**

Append to `internal/core/domain/user.go` after the `User` struct:

```go
// SearchResult is a lightweight value object for user search results.
type SearchResult struct {
	Username  string
	FirstName string
	LastName  string
}

// Name returns the concatenated full name.
func (s SearchResult) Name() string {
	return s.FirstName + " " + s.LastName
}
```

- [ ] **Step 2: Add ErrSearchQueryTooShort sentinel**

Append to `internal/core/domain/errors.go` before the final `)` in the `var` block:

```go
	// ErrSearchQueryTooShort is returned when a search query doesn't meet the minimum length (3 characters).
	ErrSearchQueryTooShort = errors.New("search query too short")
```

- [ ] **Step 3: Verify compilation**

```bash
go build ./internal/core/domain/...
```
Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add internal/core/domain/user.go internal/core/domain/errors.go
git commit -m "feat(domain): add SearchResult VO and ErrSearchQueryTooShort sentinel"
```

---

### Task 2: Add Ports (UserSearchRepository and UserUseCase extension)

**Files:**
- Modify: `internal/core/ports/user.go:31-34` (append after `UserRepository` interface, before `UserUseCase`)

**Interfaces:**
- Consumes: `domain.SearchResult` (from Task 1)
- Produces: `ports.UserSearchRepository` interface, `ports.UserUseCase.SearchUsers` method signature

- [ ] **Step 1: Add UserSearchRepository interface and extend UserUseCase**

Append to `internal/core/ports/user.go` after the `UserRepository` interface closing brace, before `UserUseCase`:

```go
// UserSearchRepository is a driven port for user search operations.
type UserSearchRepository interface {
	// Search returns users matching the query using trigram and prefix matching.
	// Results are ordered by relevance: exact match → prefix match → trigram similarity.
	Search(ctx context.Context, query string, limit int) ([]domain.SearchResult, error)
}
```

Add to the `UserUseCase` interface (inside the existing interface block, after `UpdateProfile`):

```go
	// SearchUsers searches for users by username prefix/fuzzy match.
	// query must be at least 3 characters; limit is clamped to 1–20.
	SearchUsers(ctx context.Context, query string, limit int) ([]domain.SearchResult, error)
```

- [ ] **Step 2: Verify compilation fails (missing implementation)**

```bash
go build ./...
```
Expected: compilation errors in `internal/core/usecase/` (missing `SearchUsers` method) — this confirms the interface change. Proceed to Task 3.

- [ ] **Step 3: Commit**

```bash
git add internal/core/ports/user.go
git commit -m "feat(ports): add UserSearchRepository interface and SearchUsers to UserUseCase"
```

---

### Task 3: Create Database Migration

**Files:**
- Create: `db/migrations/000013_add_user_search_indexes.up.sql`
- Create: `db/migrations/000013_add_user_search_indexes.down.sql`

**Interfaces:**
- Produces: `pg_trgm` extension enabled, `idx_users_username_prefix` (varchar_pattern_ops), `idx_users_username_trgm` (gin_trgm_ops)

- [ ] **Step 1: Create up migration**

Create `db/migrations/000013_add_user_search_indexes.up.sql`:

```sql
-- Enable trigram extension for fuzzy text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Prefix index for LIKE 'query%' patterns (optimizes leading-wildcard-free prefix matching)
CREATE INDEX idx_users_username_prefix ON users USING btree (username varchar_pattern_ops);

-- Trigram GIN index for similarity-based fuzzy matching
CREATE INDEX idx_users_username_trgm ON users USING gin (username gin_trgm_ops);
```

- [ ] **Step 2: Create down migration**

Create `db/migrations/000013_add_user_search_indexes.down.sql`:

```sql
DROP INDEX IF EXISTS idx_users_username_trgm;
DROP INDEX IF EXISTS idx_users_username_prefix;
```

- [ ] **Step 3: Run migration up to verify**

```bash
make migrate-up dsn="postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
```
Expected: migration 000013 applied successfully.

- [ ] **Step 4: Verify indexes exist in PostgreSQL**

```sql
SELECT indexname FROM pg_indexes WHERE tablename = 'users' AND indexname LIKE 'idx_users_username_%';
```
Expected: `idx_users_username_prefix` and `idx_users_username_trgm`.

- [ ] **Step 5: Commit**

```bash
git add db/migrations/000013_add_user_search_indexes.up.sql db/migrations/000013_add_user_search_indexes.down.sql
git commit -m "feat(db): add pg_trgm and varchar_pattern_ops indexes for user search"
```

---

### Task 4: Implement UserSearchRepository (GORM adapter)

**Files:**
- Create: `internal/adapters/repository/user_search_repository.go`

**Interfaces:**
- Consumes: `ports.UserSearchRepository` (from Task 2), `domain.SearchResult` (from Task 1)
- Produces: `repository.UserSearchRepository` concrete type, `NewUserSearchRepository` constructor

- [ ] **Step 1: Create the repository implementation**

Create `internal/adapters/repository/user_search_repository.go`:

```go
package repository

import (
	"context"
	"fmt"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
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
// Within each tier, results are ordered by trigram similarity descending.
func (r *userSearchRepository) Search(ctx context.Context, query string, limit int) ([]domain.SearchResult, error) {
	type row struct {
		Username  string
		FirstName string
		LastName  string
	}

	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT username, first_name, last_name
		FROM users
		WHERE username % ?
		   OR username LIKE ? || '%'
		ORDER BY
			CASE WHEN username = ? THEN 0
			     WHEN username LIKE ? || '%' THEN 1
			     ELSE 2
			END,
			similarity(username, ?) DESC
		LIMIT ?
	`, query, query, query, query, query, limit).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrRepositoryFailure, err)
	}

	results := make([]domain.SearchResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, domain.SearchResult{
			Username:  r.Username,
			FirstName: r.FirstName,
			LastName:  r.LastName,
		})
	}

	return results, nil
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./internal/adapters/repository/...
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/adapters/repository/user_search_repository.go
git commit -m "feat(repository): add UserSearchRepository with hybrid trigram/prefix SQL query"
```

---

### Task 5: Implement SearchUsers Use Case (with unit tests)

**Files:**
- Modify: `internal/core/usecase/user_usecase.go:22-24` (constructor signature)
- Modify: `internal/core/usecase/user_usecase.go:appendix` (add SearchUsers method)
- Modify: `internal/core/usecase/user_usecase_test.go:appendix` (add SearchUsers tests)

**Interfaces:**
- Consumes: `ports.UserSearchRepository` (from Task 2), `domain.ErrSearchQueryTooShort` (from Task 1), `domain.SearchResult` (from Task 1)
- Produces: `userUseCase.SearchUsers` method

- [ ] **Step 1: Write failing tests first**

Append to `internal/core/usecase/user_usecase_test.go`:

```go
func TestUserUseCase_SearchUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)
	mockSearchRepo := mocks.NewMockUserSearchRepository(ctrl)
	svc := NewUserUseCase(mockRepo, mockHasher, mockSearchRepo)

	ctx := context.Background()

	t.Run("success - returns search results", func(t *testing.T) {
		expected := []domain.SearchResult{
			{Username: "billy", FirstName: "Billy", LastName: "Kore"},
			{Username: "billie", FirstName: "Billie", LastName: "Eilish"},
		}
		mockSearchRepo.EXPECT().Search(ctx, "bil", 10).Return(expected, nil)

		results, err := svc.SearchUsers(ctx, "bil", 10)
		assert.NoError(t, err)
		assert.Equal(t, expected, results)
	})

	t.Run("success - empty results", func(t *testing.T) {
		mockSearchRepo.EXPECT().Search(ctx, "xyz", 10).Return([]domain.SearchResult{}, nil)

		results, err := svc.SearchUsers(ctx, "xyz", 10)
		assert.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("query too short - 2 chars", func(t *testing.T) {
		results, err := svc.SearchUsers(ctx, "ab", 10)
		assert.ErrorIs(t, err, domain.ErrSearchQueryTooShort)
		assert.Nil(t, results)
	})

	t.Run("query too short - empty", func(t *testing.T) {
		results, err := svc.SearchUsers(ctx, "", 10)
		assert.ErrorIs(t, err, domain.ErrSearchQueryTooShort)
		assert.Nil(t, results)
	})

	t.Run("repository error propagated", func(t *testing.T) {
		mockSearchRepo.EXPECT().Search(ctx, "bil", 10).Return(nil, domain.ErrRepositoryFailure)

		results, err := svc.SearchUsers(ctx, "bil", 10)
		assert.ErrorIs(t, err, domain.ErrRepositoryFailure)
		assert.Nil(t, results)
	})
}
```

- [ ] **Step 2: Run tests — must fail**

```bash
go test ./internal/core/usecase/ -run TestUserUseCase_SearchUsers -v
```
Expected: FAIL — `NewUserUseCase` has wrong number of arguments / `SearchUsers` undefined.

- [ ] **Step 3: Update constructor to accept UserSearchRepository**

Modify `internal/core/usecase/user_usecase.go` — update the struct and constructor:

```go
type userUseCase struct {
	userRepo       ports.UserRepository
	userSearchRepo ports.UserSearchRepository
	hasher         ports.Hasher
}

// NewUserUseCase creates a new instance of ports.UserUseCase.
func NewUserUseCase(userRepo ports.UserRepository, hasher ports.Hasher, userSearchRepo ports.UserSearchRepository) ports.UserUseCase {
	if userRepo == nil {
		panic("NewUserUseCase: userRepo is required")
	}
	if hasher == nil {
		panic("NewUserUseCase: hasher is required")
	}
	if userSearchRepo == nil {
		panic("NewUserUseCase: userSearchRepo is required")
	}
	return &userUseCase{
		userRepo:       userRepo,
		hasher:         hasher,
		userSearchRepo: userSearchRepo,
	}
}
```

- [ ] **Step 4: Add SearchUsers method**

Append after `ChangePassword` method in `internal/core/usecase/user_usecase.go`:

```go
// SearchUsers searches for users by username prefix/fuzzy match.
// query must be at least 3 characters.
func (s *userUseCase) SearchUsers(ctx context.Context, query string, limit int) ([]domain.SearchResult, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	if len(query) < 3 {
		return nil, domain.ErrSearchQueryTooShort
	}

	results, err := s.userSearchRepo.Search(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}

	return results, nil
}
```

- [ ] **Step 5: Regenerate mocks (needed for tests to compile)**

```bash
make mocks
```
Expected: generates `internal/core/ports/mocks/mock_user.go` with `MockUserSearchRepository`.

- [ ] **Step 6: Run tests — must pass**

```bash
go test ./internal/core/usecase/ -run TestUserUseCase_SearchUsers -v
```
Expected: PASS — all 5 sub-tests pass.

- [ ] **Step 7: Run all existing use case tests (regression check)**

```bash
go test ./internal/core/usecase/ -v
```
Expected: all existing tests still PASS (constructor signature changed — verify all tests compile with the new 3-arg constructor).

- [ ] **Step 8: Commit**

```bash
git add internal/core/usecase/user_usecase.go internal/core/usecase/user_usecase_test.go internal/core/ports/mocks/mock_user.go
git commit -m "feat(usecase): add SearchUsers with min-length validation and unit tests"
```

---

### Task 6: Add DTOs

**Files:**
- Modify: `internal/api/dto/user_dto.go:appendix`

**Interfaces:**
- Produces: `dto.SearchUsersRequest`, `dto.SearchUsersItem`, `dto.SearchUsersResponse`

- [ ] **Step 1: Add search DTOs**

Append to `internal/api/dto/user_dto.go`:

```go
// SearchUsersRequest holds query parameters for the user search endpoint.
type SearchUsersRequest struct {
	Q     string `query:"q" validate:"required,min=3"`
	Limit int    `query:"limit" validate:"omitempty,min=1,max=20"`
}

// SearchUsersItem is a single search result item.
type SearchUsersItem struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

// SearchUsersResponse wraps the search results.
type SearchUsersResponse struct {
	Data []SearchUsersItem `json:"data"`
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./internal/api/dto/...
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/api/dto/user_dto.go
git commit -m "feat(dto): add SearchUsersRequest, SearchUsersItem, SearchUsersResponse"
```

---

### Task 7: Add SearchUsers Handler (with unit tests)

**Files:**
- Modify: `internal/api/handler/user_handler.go:appendix` (add `SearchUsers` method)
- Create: `internal/api/handler/user_handler_test.go` (or modify existing if present)

**Interfaces:**
- Consumes: `ports.UserUseCase.SearchUsers` (from Task 2/5), `dto.SearchUsersRequest` (from Task 6)
- Produces: `UserHandler.SearchUsers` handler method

- [ ] **Step 1: Check if handler test file exists**

```bash
ls internal/api/handler/*_test.go
```

- [ ] **Step 2: Add SearchUsers handler method**

Append to `internal/api/handler/user_handler.go` (before the `toUserResponse` helper):

```go
// SearchUsers handles the GET /users/search endpoint.
//
//	@Summary		Search users
//	@Description	Search users by username using prefix and fuzzy matching.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			q		query		string	true	"Search query (min 3 characters)"
//	@Param			limit	query		int		false	"Max results (1-20, default 10)"
//	@Success		200		{object}	dto.SearchUsersResponse
//	@Failure		400		{object}	dto.ProblemDetail
//	@Failure		500		{object}	dto.ProblemDetail
//	@Router			/users/search [get]
func (h *UserHandler) SearchUsers(c echo.Context) error {
	var req dto.SearchUsersRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error(c.Request().Context(), "SearchUsers failed", "error", "Invalid query parameters")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error(c.Request().Context(), "SearchUsers failed", "validation_error", err)
		return err
	}

	results, err := h.userUseCase.SearchUsers(c.Request().Context(), req.Q, req.Limit)
	if err != nil {
		h.log.Error(c.Request().Context(), "SearchUsers failed", "query", req.Q, "error", err)
		return err
	}

	items := make([]dto.SearchUsersItem, 0, len(results))
	for _, r := range results {
		items = append(items, dto.SearchUsersItem{
			Username: r.Username,
			Name:     r.Name(),
		})
	}

	h.log.Info(c.Request().Context(), "SearchUsers succeeded", "query", req.Q, "results", len(items))
	return c.JSON(http.StatusOK, dto.SearchUsersResponse{Data: items})
}
```

- [ ] **Step 3: Register error mapping**

Append to the `errorMappings` map in `internal/api/middleware/error_registry.go`:

```go
	domain.ErrSearchQueryTooShort: {http.StatusBadRequest, domain.CodeInvalidArgument, "invalid-argument", "Bad Request", "Search query must be at least 3 characters"},
```

- [ ] **Step 4: Verify compilation before writing tests**

```bash
go build ./internal/api/handler/...
```
Expected: no errors.

- [ ] **Step 5: Write handler unit test**

If `internal/api/handler/user_handler_test.go` does not exist, create it. Otherwise append:

```go
package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_SearchUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	mockLoginUC := mocks.NewMockLoginUseCase(ctrl)
	mockFollowUC := mocks.NewMockFollowUseCase(ctrl)
	mockPostUC := mocks.NewMockPostUseCase(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	// Logger expectations for any calls
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	h := NewUserHandler(mockUserUC, mockLoginUC, mockFollowUC, mockPostUC, mockValidator, mockLogger)

	t.Run("success with results", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/users/search?q=bil&limit=10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
		mockUserUC.EXPECT().SearchUsers(gomock.Any(), "bil", 10).Return([]domain.SearchResult{
			{Username: "billy", FirstName: "Billy", LastName: "Kore"},
		}, nil)

		err := h.SearchUsers(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"username":"billy"`)
	})

	t.Run("success with empty results", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/users/search?q=xyz", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
		mockUserUC.EXPECT().SearchUsers(gomock.Any(), "xyz", 10).Return([]domain.SearchResult{}, nil)

		err := h.SearchUsers(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"data":[]`)
	})

	t.Run("default limit when not provided", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/users/search?q=bil", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
		mockUserUC.EXPECT().SearchUsers(gomock.Any(), "bil", 10).Return([]domain.SearchResult{}, nil)

		err := h.SearchUsers(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
```

- [ ] **Step 6: Run handler tests**

```bash
go test ./internal/api/handler/ -run TestUserHandler_SearchUsers -v
```
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/api/handler/user_handler.go internal/api/handler/user_handler_test.go internal/api/middleware/error_registry.go
git commit -m "feat(handler): add SearchUsers endpoint handler with tests and error mapping"
```

---

### Task 8: Wire Everything in main.go

**Files:**
- Modify: `cmd/main.go:131-140` (constructor area — add searchRepo)
- Modify: `cmd/main.go:141` (update NewUserUseCase call)
- Modify: `cmd/main.go:184-186` (register route)

**Interfaces:**
- Consumes: `repository.NewUserSearchRepository` (from Task 4), updated `usecase.NewUserUseCase` (from Task 5)
- Produces: Wired application

- [ ] **Step 1: Create UserSearchRepository instance**

In `cmd/main.go`, in `newApplication()`, add after `userRepo := repository.NewUserRepository(db)`:

```go
	userSearchRepo := repository.NewUserSearchRepository(db)
```

- [ ] **Step 2: Pass UserSearchRepository to NewUserUseCase**

Change the line:
```go
	userUc := usecase.NewUserUseCase(userRepo, hasherSvc)
```
to:
```go
	userUc := usecase.NewUserUseCase(userRepo, hasherSvc, userSearchRepo)
```

- [ ] **Step 3: Register the route**

In the `registerRoutes` function, add after the public `users.GET("/:username", userHdl.GetUser)` line:

```go
	users.GET("/search", userHdl.SearchUsers)
```

This must be placed **before** `users.GET("/:username", ...)` so the literal `/search` path is matched before the `:username` wildcard.

So the final order in the public users group should be:
```go
	users.GET("/search", userHdl.SearchUsers)
	users.GET("/:username", userHdl.GetUser)
	users.GET("/:username/posts", userHdl.GetUserPosts)
```

- [ ] **Step 4: Verify compilation**

```bash
go build ./cmd/...
```
Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add cmd/main.go
git commit -m "feat(main): wire UserSearchRepository and register GET /users/search route"
```

---

### Task 9: Regenerate Swagger Docs

**Files:**
- Regenerate: `api/swagger/swagger.json`
- Regenerate: `api/swagger/swagger.yaml`
- Regenerate: `api/swagger/docs.go`

- [ ] **Step 1: Regenerate swagger docs**

```bash
make docs
```
Expected: Swagger docs regenerated successfully, new `/users/search` endpoint appears.

- [ ] **Step 2: Verify the new endpoint appears in swagger.yaml**

```bash
grep -A5 "/users/search" api/swagger/swagger.yaml
```
Expected: shows the search endpoint definition.

- [ ] **Step 3: Commit**

```bash
git add api/swagger/
git commit -m "docs(swagger): regenerate with user search endpoint"
```

---

### Task 10: Run Full Test Suite

**Files:**
- All modified files

- [ ] **Step 1: Run all tests**

```bash
make test
```
Expected: all tests PASS, no regressions.

- [ ] **Step 2: Run linter**

```bash
make lint
```
Expected: no new lint errors.

- [ ] **Step 3: Run go vet**

```bash
make vet
```
Expected: no vet warnings.

---

### Task 11: Manual Verification (Optional)

- [ ] **Step 1: Start the server**

```bash
make run
```

- [ ] **Step 2: Test exact match**

```bash
curl "http://localhost:8080/users/search?q=billy&limit=5" | jq
```
Expected: `billy` appears first in results.

- [ ] **Step 3: Test prefix match**

```bash
curl "http://localhost:8080/users/search?q=bil&limit=5" | jq
```
Expected: usernames starting with `bil` appear before fuzzy matches.

- [ ] **Step 4: Test query too short**

```bash
curl -s "http://localhost:8080/users/search?q=ab" | jq
```
Expected: 400 with `"code": "INVALID_ARGUMENT"`.

- [ ] **Step 5: Test no results**

```bash
curl "http://localhost:8080/users/search?q=zzzzzzzz" | jq
```
Expected: 200 with `"data": []`.

- [ ] **Step 6: Test default limit**

```bash
curl "http://localhost:8080/users/search?q=bil" | jq '.data | length'
```
Expected: ≤ 10 results.

---

## Final Checklist

- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] `make vet` passes
- [ ] `make docs` regenerates successfully
- [ ] Migration 000013 applies and rolls back cleanly
- [ ] New route appears before `/:username` wildcard in route registration
