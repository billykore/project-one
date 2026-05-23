# Post Comments in GET Post Endpoint Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Include post comments in the response of the `GET /posts/{id}` endpoint.

**Architecture:** Use Clean Architecture principles. The `PostUseCase` will orchestrate the fetching of the post from the `PostRepository` and the retrieval of its associated comments from the `CommentRepository`, combining them into a modified `domain.Post` entity. This keeps repositories decoupled, prevents unnecessary comment preloading during updates/deletes, and ensures testability.

**Tech Stack:** Go 1.26+, Echo, GORM, GoMock, Testify, Swaggo.

---

## Design Decisions, Tradeoffs, and Edge Cases

### 1. Usecase Orchestration vs Repository Preloading

* **Decision:** We will fetch comments using a separate `GetByPostID` query inside `CommentRepository`, orchestrated by the `PostUseCase`.
* **Tradeoff:** This introduces an additional database round-trip (2 queries instead of 1 join). However, this is preferred over GORM preloading because:
  * **Decoupling:** `PostRepository` remains responsible only for posts, and `CommentRepository` only for comments.
  * **Efficiency:** Usecase methods like `UpdatePost` or `DeletePost` call `repo.GetByID` to verify existence/ownership. Preloading comments on all `GetByID` calls would cause unnecessary overhead for update/delete operations.
  * **Future-Proofing:** It makes it straightforward to add pagination, limits, or filters to comments in the future (e.g. only returning the first 10 comments), which is difficult to optimize with GORM preloads.
* **Mocks/Testing:** Decoupled repositories make unit testing in `post_usecase_test.go` clean and straightforward using mocks.

### 2. Ordering of Comments

* **Decision:** Comments will be retrieved in chronological order (`created_at ASC` or `id ASC`) to allow reading comments in order of posting.

### 3. Edge Cases

* **Post with zero comments:** The comment repository query will return an empty slice. The API handler must return an empty JSON array `[]` instead of `null`.
* **Soft-deleted comments:** GORM automatically filters out soft-deleted comments (where `deleted_at IS NULL`) since `commentModel` uses `gorm.Model`.
* **Comments query failure:** If the database query for comments fails, the usecase logs the error and returns `domain.ErrInternalServer`. The handler will convert this into a HTTP 500 error.
* **Authentication/Permissions:** The endpoint already requires authentication and ownership validation, which is preserved.

---

## Step-by-Step Implementation Tasks

### Task 1: Update Domain Entity

**Files:**

* Modify: `internal/core/domain/post.go:15-26`

* [ ] **Step 1: Add the Comments field to the domain.Post struct**

Edit `internal/core/domain/post.go` to include `Comments []*Comment`.

```go
// Post is the core domain entity representing a user's post.
type Post struct {
 ID        int
 Username  string
 Title     string
 Content   string
 Tags      []string
 Comments  []*Comment
 CreatedAt time.Time
 UpdatedAt time.Time
 DeletedAt time.Time
}
```

* [ ] **Step 2: Verify compile passes**
Run: `go vet ./internal/core/domain/...`
Expected: Success with no errors.

---

### Task 2: Update Ports Layer

**Files:**

* Modify: `internal/core/ports/comment.go:9-13`

* [ ] **Step 1: Add GetByPostID to ports.CommentRepository**

Edit `internal/core/ports/comment.go` to define the method signature.

```go
// CommentRepository is a driven port for comment persistence.
type CommentRepository interface {
 // Create saves a new comment to the repository.
 Create(ctx context.Context, comment *domain.Comment) error
 // GetByPostID retrieves all comments for a specific post.
 GetByPostID(ctx context.Context, postID int) ([]*domain.Comment, error)
}
```

* [ ] **Step 2: Verify compile passes**
Run: `go vet ./internal/core/ports/...`
Expected: Success with no errors.

---

### Task 3: Regenerate Mock Files

**Files:**

* Modify: `internal/core/usecase/mocks/mock_comment.go` (automatic regeneration)

* [ ] **Step 1: Run mock generation script**
Run: `make mock`
Expected: `./scripts/mock.sh` runs successfully, regenerating the mock files including `mock_comment.go` which now includes the mocked `GetByPostID` method.

---

### Task 4: Implement Comment Repository Method

**Files:**

* Modify: `internal/adapters/repository/comment_repository.go:50-59`

* [ ] **Step 1: Implement GetByPostID in commentRepository**

Edit `internal/adapters/repository/comment_repository.go` to implement the new port method using GORM. The query must sort comments by `created_at ASC` and automatically filters out soft-deleted comments (handled by GORM).

```go
func (r *commentRepository) GetByPostID(ctx context.Context, postID int) ([]*domain.Comment, error) {
 var models []commentModel
 err := r.db.WithContext(ctx).
  Where("post_id = ?", postID).
  Order("created_at ASC").
  Find(&models).Error
 if err != nil {
  return nil, err
 }

 comments := make([]*domain.Comment, 0, len(models))
 for _, m := range models {
  comments = append(comments, m.toDomain())
 }
 return comments, nil
}
```

* [ ] **Step 2: Verify compile passes**
Run: `go vet ./internal/adapters/repository/...`
Expected: Success with no errors.

---

### Task 5: Update Post UseCase

**Files:**

* Modify: `internal/core/usecase/post_usecase.go:12-63`

* [ ] **Step 1: Update postUseCase struct and constructor to accept CommentRepository**

Inject `commentRepo ports.CommentRepository` into `NewPostUseCase` and add it to the `postUseCase` struct.

```go
type postUseCase struct {
 repo        ports.PostRepository
 commentRepo ports.CommentRepository
 log         ports.Logger
}

// NewPostUseCase creates a new instance of ports.PostUseCase.
func NewPostUseCase(repo ports.PostRepository, commentRepo ports.CommentRepository, log ports.Logger) ports.PostUseCase {
 if repo == nil {
  panic("NewPostUseCase: repo is required")
 }
 if commentRepo == nil {
  panic("NewPostUseCase: commentRepo is required")
 }
 if log == nil {
  panic("NewPostUseCase: log is required")
 }
 return &postUseCase{
  repo:        repo,
  commentRepo: commentRepo,
  log:         log,
 }
}
```

* [ ] **Step 2: Update GetPostByID to fetch and attach comments**

Modify `GetPostByID` in `internal/core/usecase/post_usecase.go` to fetch comments and set them on `post.Comments`.

```go
func (s *postUseCase) GetPostByID(ctx context.Context, username string, id int) (*domain.Post, error) {
 if id <= 0 {
  return nil, domain.ErrInvalidPost
 }

 post, err := s.repo.GetByID(ctx, username, id)
 if err != nil {
  if errors.Is(err, domain.ErrPostNotFound) {
   return nil, err
  }
  s.log.Error(ctx, "failed to get post by id", "postID", id, "username", username, "error", err)
  return nil, domain.ErrInternalServer
 }

 comments, err := s.commentRepo.GetByPostID(ctx, id)
 if err != nil {
  s.log.Error(ctx, "failed to get comments for post", "postID", id, "error", err)
  return nil, domain.ErrInternalServer
 }
 post.Comments = comments

 return post, nil
}
```

* [ ] **Step 3: Verify compile passes**
Run: `go vet ./internal/core/usecase/...`
Expected: Success with no errors.

---

### Task 6: Update Dependency Injection (cmd/main.go)

**Files:**

* Modify: `cmd/main.go:76`

* [ ] **Step 1: Inject commentRepo into NewPostUseCase**

Edit `cmd/main.go` to pass `commentRepo` to the usecase constructor.

```go
 postUc := usecase.NewPostUseCase(postRepo, commentRepo, lgr)
```

* [ ] **Step 2: Verify compile passes**
Run: `go vet ./cmd/main.go`
Expected: Success with no errors.

---

### Task 7: Update Post Response DTOs

**Files:**

* Modify: `internal/api/dto/comment_dto.go:1-7`
* Modify: `internal/api/dto/post_dto.go:16-25`

* [ ] **Step 1: Define CommentResponse in comment_dto.go**

Open `internal/api/dto/comment_dto.go` and append the `CommentResponse` struct. Ensure `"time"` is imported.

```go
package dto

import "time"

type CreateCommentRequest struct {
 ID      int    `param:"id" validate:"required,min=1"`
 Content string `json:"content" validate:"required,min=1"`
}

type CommentResponse struct {
 ID        int       `json:"id"`
 Username  string    `json:"username"`
 Content   string    `json:"content"`
 CreatedAt time.Time `json:"created_at"`
}
```

* [ ] **Step 2: Add Comments field to PostResponse DTO**

Open `internal/api/dto/post_dto.go` and update `PostResponse` to include the comments slice.

```go
type PostResponse struct {
 ID        int                `json:"id"`
 Message   string             `json:"message,omitempty"`
 Title     string             `json:"title,omitempty"`
 Content   string             `json:"content,omitempty"`
 Tags      []string           `json:"tags,omitempty"`
 Author    string             `json:"author,omitempty"`
 Comments  []*CommentResponse `json:"comments,omitempty"`
 CreatedAt time.Time          `json:"created_at"`
 UpdatedAt time.Time          `json:"updated_at"`
}
```

* [ ] **Step 3: Verify compile passes**
Run: `go vet ./internal/api/dto/...`
Expected: Success with no errors.

---

### Task 8: Update Post Handler

**Files:**

* Modify: `internal/api/handler/post_handler.go:100-134`

* [ ] **Step 1: Update GetPostByID to map comments to PostResponse**

Modify the response mapping in `GetPostByID` within `internal/api/handler/post_handler.go`. Make sure to initialize the slice to `[]*dto.CommentResponse{}` when `post.Comments` is nil/empty so the JSON response returns `[]` instead of `null`.

```go
 commentsResp := make([]*dto.CommentResponse, 0)
 for _, comment := range post.Comments {
  commentsResp = append(commentsResp, &dto.CommentResponse{
   ID:        comment.ID,
   Username:  comment.Username,
   Content:   comment.Content,
   CreatedAt: comment.CreatedAt,
  })
 }

 return c.JSON(http.StatusOK, dto.PostResponse{
  ID:        post.ID,
  Title:     post.Title,
  Content:   post.Content,
  Tags:      post.Tags,
  Comments:  commentsResp,
  CreatedAt: post.CreatedAt,
  UpdatedAt: post.UpdatedAt,
 })
```

* [ ] **Step 2: Verify compile passes**
Run: `go vet ./internal/api/handler/...`
Expected: Success with no errors.

---

### Task 9: Update Unit Tests

**Files:**

* Modify: `internal/core/usecase/post_usecase_test.go:14-288`

* [ ] **Step 1: Inject mockCommentRepo into all test cases**

Update `TestPostUseCase_CreatePost`, `TestPostUseCase_GetPostByID`, `TestPostUseCase_GetPosts`, `TestPostUseCase_UpdatePost`, and `TestPostUseCase_DeletePost` to instantiate and inject the mock `CommentRepository`.

For example, in `TestPostUseCase_CreatePost`:

```go
 mockRepo := mocks.NewMockPostRepository(ctrl)
 mockCommentRepo := mocks.NewMockCommentRepository(ctrl)
 mockLog := mocks.NewMockLogger(ctrl)
 svc := NewPostUseCase(mockRepo, mockCommentRepo, mockLog)
```

* [ ] **Step 2: Add comments expectations to TestPostUseCase_GetPostByID success scenario**

In `TestPostUseCase_GetPostByID`, mock the `GetByPostID` method of `CommentRepository` to return comments.

```go
 t.Run("success", func(t *testing.T) {
  expectedPost := &domain.Post{ID: postID, Username: username, Title: "Test Title"}
  expectedComments := []*domain.Comment{
   {ID: 1, PostID: postID, Username: "commenter", Content: "Nice post"},
  }
  mockRepo.EXPECT().GetByID(ctx, username, postID).Return(expectedPost, nil)
  mockCommentRepo.EXPECT().GetByPostID(ctx, postID).Return(expectedComments, nil)

  post, err := svc.GetPostByID(ctx, username, postID)

  assert.NoError(t, err)
  assert.Equal(t, expectedPost, post)
  assert.Equal(t, expectedComments, post.Comments)
 })
```

* [ ] **Step 3: Add comments database error test case**

In `TestPostUseCase_GetPostByID`, add a test case to cover when comment fetching fails.

```go
 t.Run("comment repository error", func(t *testing.T) {
  expectedPost := &domain.Post{ID: postID, Username: username, Title: "Test Title"}
  mockRepo.EXPECT().GetByID(ctx, username, postID).Return(expectedPost, nil)
  mockCommentRepo.EXPECT().GetByPostID(ctx, postID).Return(nil, errors.New("db error"))
  mockLog.EXPECT().Error(ctx, "failed to get comments for post", "postID", postID, "error", gomock.Any())

  post, err := svc.GetPostByID(ctx, username, postID)

  assert.Error(t, err)
  assert.Nil(t, post)
  assert.True(t, errors.Is(err, domain.ErrInternalServer))
 })
```

* [ ] **Step 4: Verify all unit tests pass**
Run: `make test`
Expected: All tests pass.

---

### Task 10: Verify and Regenerate Swagger Docs

* [ ] **Step 1: Regenerate Swagger Documentation**
Run: `make docs`
Expected: Swaggo successfully parses updated comments and structs, updating files under `api/swagger/`.

* [ ] **Step 2: Run all workspace checks**
Run: `make check`
Expected: All tests, linting, docs compilation, and code vetting pass cleanly.
