# Idempotent Likes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor post likes/unlikes to be idempotent by splitting ToggleLike into distinct POST and DELETE /posts/:id/likes APIs.

**Architecture:** Use Clean Architecture patterns. We'll update the driving port `PostUseCase` with two distinct methods, implement them in the usecase layer using existing database repository ports, and update handler and routing.

**Tech Stack:** Go 1.26+, Echo framework, GORM, testify/mock for unit testing.

---

### Task 1: Port Interface and Mock Update

**Files:**
- Modify: `internal/core/ports/post.go`
- Modify: `internal/core/usecase/mocks/mock_like.go`

- [ ] **Step 1: Modify PostUseCase interface**
  Update the `PostUseCase` interface to replace `ToggleLike` with `LikePost` and `UnlikePost`.
  
  In `internal/core/ports/post.go`:
  ```go
  type PostUseCase interface {
  	CreatePost(ctx context.Context, username string, title, content string, tags []string) (*domain.Post, error)
  	GetPostByID(ctx context.Context, username string, postID int) (*domain.Post, error)
  	GetPosts(ctx context.Context, username string, limit, offset int) ([]*domain.Post, error)
  	UpdatePost(ctx context.Context, username string, postID int, title, content string) (*domain.Post, error)
  	DeletePost(ctx context.Context, username string, postID int) error
  	// LikePost likes a post by the given username. If already liked, it behaves idempotently.
  	LikePost(ctx context.Context, postID int, username string) (likeCount int, err error)
  	// UnlikePost unlikes a post by the given username. If not liked, it behaves idempotently.
  	UnlikePost(ctx context.Context, postID int, username string) (likeCount int, err error)
  	GetLikeStatus(ctx context.Context, postID int, username string) (liked bool, likeCount int, err error)
  }
  ```

- [ ] **Step 2: Regenerate mocks**
  Run: `make mock`
  Expected output: Success output regenerating mock files.

- [ ] **Step 3: Commit interface changes**
  Run: `git add internal/core/ports/post.go internal/core/usecase/mocks/mock_like.go && git commit -m "refactor: split ToggleLike into LikePost and UnlikePost in PostUseCase port"`

---

### Task 2: Implement LikePost in UseCase

**Files:**
- Modify: `internal/core/usecase/post_usecase.go`
- Modify: `internal/core/usecase/post_usecase_test.go`

- [ ] **Step 1: Write failing tests for LikePost**
  Add unit tests in `internal/core/usecase/post_usecase_test.go`.
  
  Add to `internal/core/usecase/post_usecase_test.go`:
  ```go
  func TestPostUseCase_LikePost(t *testing.T) {
  	ctrl := gomock.NewController(t)
  	defer ctrl.Finish()

  	mockRepo := mocks.NewMockPostRepository(ctrl)
  	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
  	mockLog := mocks.NewMockLogger(ctrl)
  	svc := NewPostUseCase(mockRepo, mockLikeRepo, mockLog)

  	ctx := context.Background()
  	username := "testuser"
  	postID := 1

  	t.Run("success - new like", func(t *testing.T) {
  		mockLikeRepo.EXPECT().Exists(ctx, postID, username).Return(false, nil)
  		mockLikeRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
  		mockRepo.EXPECT().IncrementLikeCount(ctx, postID, 1).Return(nil)
  		mockRepo.EXPECT().GetByIDOnly(ctx, postID).Return(&domain.Post{ID: postID, LikeCount: 5}, nil)
  		mockLog.EXPECT().Info(ctx, "post liked successfully", "postID", postID, "username", username)

  		count, err := svc.LikePost(ctx, postID, username)
  		assert.NoError(t, err)
  		assert.Equal(t, 5, count)
  	})

  	t.Run("success idempotent - already liked", func(t *testing.T) {
  		mockLikeRepo.EXPECT().Exists(ctx, postID, username).Return(true, nil)
  		mockRepo.EXPECT().GetByIDOnly(ctx, postID).Return(&domain.Post{ID: postID, LikeCount: 4}, nil)

  		count, err := svc.LikePost(ctx, postID, username)
  		assert.NoError(t, err)
  		assert.Equal(t, 4, count)
  	})

  	t.Run("post not found", func(t *testing.T) {
  		mockLikeRepo.EXPECT().Exists(ctx, postID, username).Return(false, nil)
  		mockLikeRepo.EXPECT().Create(ctx, gomock.Any()).Return(domain.ErrPostNotFound)

  		count, err := svc.LikePost(ctx, postID, username)
  		assert.Error(t, err)
  		assert.True(t, errors.Is(err, domain.ErrPostNotFound))
  		assert.Equal(t, 0, count)
  	})
  }
  ```

- [ ] **Step 2: Run test and verify it fails to compile/run**
  Run: `go test ./internal/core/usecase`
  Expected output: Compilation failure because `LikePost` is not implemented on `*postUseCase`.

- [ ] **Step 3: Implement LikePost in post_usecase.go**
  Replace `ToggleLike` in `internal/core/usecase/post_usecase.go` with `LikePost` (and we will also add a stub for `UnlikePost` to make it compile).
  
  In `internal/core/usecase/post_usecase.go`:
  ```go
  func (uc *postUseCase) LikePost(ctx context.Context, postID int, username string) (int, error) {
  	if postID <= 0 {
  		return 0, domain.ErrInvalidPost
  	}
  	if username == "" {
  		return 0, domain.ErrValidationFailed
  	}

  	// Check if already liked (idempotency check)
  	exists, err := uc.likeRepo.Exists(ctx, postID, username)
  	if err != nil {
  		uc.log.Error(ctx, "failed to check if like exists", "postID", postID, "username", username, "error", err)
  		return 0, domain.ErrInternalServer
  	}
  	if exists {
  		post, err := uc.postRepo.GetByIDOnly(ctx, postID)
  		if err != nil {
  			uc.log.Error(ctx, "failed to get post for like count", "postID", postID, "error", err)
  			return 0, domain.ErrInternalServer
  		}
  		return post.LikeCount, nil
  	}

  	like := &domain.Like{
  		PostID:   postID,
  		Username: username,
  	}
  	if err := uc.likeRepo.Create(ctx, like); err != nil {
  		if errors.Is(err, domain.ErrPostNotFound) {
  			return 0, err
  		}
  		uc.log.Error(ctx, "failed to create like", "postID", postID, "username", username, "error", err)
  		return 0, domain.ErrInternalServer
  	}

  	if err := uc.postRepo.IncrementLikeCount(ctx, postID, 1); err != nil {
  		uc.log.Error(ctx, "failed to increment like count", "postID", postID, "error", err)
  	}
  	uc.log.Info(ctx, "post liked successfully", "postID", postID, "username", username)

  	post, err := uc.postRepo.GetByIDOnly(ctx, postID)
  	if err != nil {
  		uc.log.Error(ctx, "failed to get post for like count", "postID", postID, "error", err)
  		return 0, domain.ErrInternalServer
  	}

  	return post.LikeCount, nil
  }

  func (uc *postUseCase) UnlikePost(ctx context.Context, postID int, username string) (int, error) {
  	return 0, nil // Stub
  }
  ```

- [ ] **Step 4: Run tests and verify they pass**
  Run: `go test -v ./internal/core/usecase -run TestPostUseCase_LikePost`
  Expected output: PASS

- [ ] **Step 5: Commit LikePost implementation**
  Run: `git add internal/core/usecase/post_usecase.go internal/core/usecase/post_usecase_test.go && git commit -m "feat: implement idempotent LikePost in usecase"`

---

### Task 3: Implement UnlikePost in UseCase

**Files:**
- Modify: `internal/core/usecase/post_usecase.go`
- Modify: `internal/core/usecase/post_usecase_test.go`

- [ ] **Step 1: Write failing tests for UnlikePost**
  Add unit tests in `internal/core/usecase/post_usecase_test.go`.
  
  Add to `internal/core/usecase/post_usecase_test.go`:
  ```go
  func TestPostUseCase_UnlikePost(t *testing.T) {
  	ctrl := gomock.NewController(t)
  	defer ctrl.Finish()

  	mockRepo := mocks.NewMockPostRepository(ctrl)
  	mockLikeRepo := mocks.NewMockLikeRepository(ctrl)
  	mockLog := mocks.NewMockLogger(ctrl)
  	svc := NewPostUseCase(mockRepo, mockLikeRepo, mockLog)

  	ctx := context.Background()
  	username := "testuser"
  	postID := 1

  	t.Run("success - unlike existing", func(t *testing.T) {
  		mockLikeRepo.EXPECT().Exists(ctx, postID, username).Return(true, nil)
  		mockLikeRepo.EXPECT().Delete(ctx, postID, username).Return(nil)
  		mockRepo.EXPECT().IncrementLikeCount(ctx, postID, -1).Return(nil)
  		mockRepo.EXPECT().GetByIDOnly(ctx, postID).Return(&domain.Post{ID: postID, LikeCount: 3}, nil)
  		mockLog.EXPECT().Info(ctx, "post unliked successfully", "postID", postID, "username", username)

  		count, err := svc.UnlikePost(ctx, postID, username)
  		assert.NoError(t, err)
  		assert.Equal(t, 3, count)
  	})

  	t.Run("success idempotent - not liked", func(t *testing.T) {
  		mockLikeRepo.EXPECT().Exists(ctx, postID, username).Return(false, nil)
  		mockRepo.EXPECT().GetByIDOnly(ctx, postID).Return(&domain.Post{ID: postID, LikeCount: 4}, nil)

  		count, err := svc.UnlikePost(ctx, postID, username)
  		assert.NoError(t, err)
  		assert.Equal(t, 4, count)
  	})
  }
  ```

- [ ] **Step 2: Run tests to verify the stub fails**
  Run: `go test -v ./internal/core/usecase -run TestPostUseCase_UnlikePost`
  Expected output: FAIL (returned count 0 instead of expected)

- [ ] **Step 3: Implement UnlikePost in post_usecase.go**
  Replace the `UnlikePost` stub.
  
  In `internal/core/usecase/post_usecase.go`:
  ```go
  func (uc *postUseCase) UnlikePost(ctx context.Context, postID int, username string) (int, error) {
  	if postID <= 0 {
  		return 0, domain.ErrInvalidPost
  	}
  	if username == "" {
  		return 0, domain.ErrValidationFailed
  	}

  	// Check if like exists (idempotency check)
  	exists, err := uc.likeRepo.Exists(ctx, postID, username)
  	if err != nil {
  		uc.log.Error(ctx, "failed to check if like exists", "postID", postID, "username", username, "error", err)
  		return 0, domain.ErrInternalServer
  	}
  	if !exists {
  		post, err := uc.postRepo.GetByIDOnly(ctx, postID)
  		if err != nil {
  			uc.log.Error(ctx, "failed to get post for like count", "postID", postID, "error", err)
  			return 0, domain.ErrInternalServer
  		}
  		return post.LikeCount, nil
  	}

  	if err := uc.likeRepo.Delete(ctx, postID, username); err != nil {
  		uc.log.Error(ctx, "failed to delete like", "postID", postID, "username", username, "error", err)
  		return 0, domain.ErrInternalServer
  	}

  	if err := uc.postRepo.IncrementLikeCount(ctx, postID, -1); err != nil {
  		uc.log.Error(ctx, "failed to decrement like count", "postID", postID, "error", err)
  	}
  	uc.log.Info(ctx, "post unliked successfully", "postID", postID, "username", username)

  	post, err := uc.postRepo.GetByIDOnly(ctx, postID)
  	if err != nil {
  		uc.log.Error(ctx, "failed to get post for like count", "postID", postID, "error", err)
  		return 0, domain.ErrInternalServer
  	}

  	return post.LikeCount, nil
  }
  ```

- [ ] **Step 4: Run tests and verify they pass**
  Run: `go test -v ./internal/core/usecase -run TestPostUseCase_UnlikePost`
  Expected output: PASS

- [ ] **Step 5: Run all usecase tests**
  Run: `go test ./internal/core/usecase`
  Expected output: PASS

- [ ] **Step 6: Commit UnlikePost implementation**
  Run: `git add internal/core/usecase/post_usecase.go internal/core/usecase/post_usecase_test.go && git commit -m "feat: implement idempotent UnlikePost in usecase"`

---

### Task 4: Refactor Handler and Routes

**Files:**
- Modify: `internal/api/handler/post_handler.go`
- Modify: `cmd/main.go`

- [ ] **Step 1: Implement LikePost and UnlikePost in post_handler.go**
  Replace `ToggleLike` handler with `LikePost` and `UnlikePost` handlers in `internal/api/handler/post_handler.go`. Update Swagger comments.
  
  In `internal/api/handler/post_handler.go`:
  ```go
  // LikePost handles the POST /posts/:id/likes endpoint.
  //
  //	@Summary		Like post
  //	@Description	Like a post idempotently.
  //	@Tags			posts,likes
  //	@Produce		json
  //	@Param			id	path		int	true	"Post ID"
  //	@Success		200	{object}	dto.LikeResponse
  //	@Failure		400	{object}	dto.ErrorResponse
  //	@Failure		401	{object}	dto.ErrorResponse
  //	@Failure		404	{object}	dto.ErrorResponse
  //	@Failure		500	{object}	dto.ErrorResponse
  //	@Security		BearerAuth
  //	@Router			/posts/{id}/likes [post]
  func (h *PostHandler) LikePost(c echo.Context) error {
  	username, ok := c.Get("username").(string)
  	if !ok {
  		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
  	}

  	idStr := c.Param("id")
  	id, err := strconv.Atoi(idStr)
  	if err != nil {
  		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Post ID must be a number"})
  	}

  	likeCount, err := h.postUseCase.LikePost(c.Request().Context(), id, username)
  	if err != nil {
  		if errors.Is(err, domain.ErrPostNotFound) {
  			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Post not found"})
  		}
  		if errors.Is(err, domain.ErrInvalidPost) {
  			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Post ID must be integer and not 0"})
  		}
  		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
  	}

  	return c.JSON(http.StatusOK, dto.LikeResponse{
  		Liked:     true,
  		LikeCount: likeCount,
  	})
  }

  // UnlikePost handles the DELETE /posts/:id/likes endpoint.
  //
  //	@Summary		Unlike post
  //	@Description	Unlike a post idempotently.
  //	@Tags			posts,likes
  //	@Produce		json
  //	@Param			id	path		int	true	"Post ID"
  //	@Success		200	{object}	dto.LikeResponse
  //	@Failure		400	{object}	dto.ErrorResponse
  //	@Failure		401	{object}	dto.ErrorResponse
  //	@Failure		404	{object}	dto.ErrorResponse
  //	@Failure		500	{object}	dto.ErrorResponse
  //	@Security		BearerAuth
  //	@Router			/posts/{id}/likes [delete]
  func (h *PostHandler) UnlikePost(c echo.Context) error {
  	username, ok := c.Get("username").(string)
  	if !ok {
  		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
  	}

  	idStr := c.Param("id")
  	id, err := strconv.Atoi(idStr)
  	if err != nil {
  		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Post ID must be a number"})
  	}

  	likeCount, err := h.postUseCase.UnlikePost(c.Request().Context(), id, username)
  	if err != nil {
  		if errors.Is(err, domain.ErrPostNotFound) {
  			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Post not found"})
  		}
  		if errors.Is(err, domain.ErrInvalidPost) {
  			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Post ID must be integer and not 0"})
  		}
  		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Something went wrong"})
  	}

  	return c.JSON(http.StatusOK, dto.LikeResponse{
  		Liked:     false,
  		LikeCount: likeCount,
  	})
  }
  ```

- [ ] **Step 2: Update route registration in main.go**
  Replace `ToggleLike` route with `LikePost` and `UnlikePost` routes in `cmd/main.go`.
  
  In `cmd/main.go`:
  ```go
  		// Likes on posts.
  		posts.POST("/:id/likes", postHdl.LikePost)
  		posts.DELETE("/:id/likes", postHdl.UnlikePost)
  		posts.GET("/:id/likes", postHdl.GetLikeStatus)
  ```

- [ ] **Step 3: Regenerate Swagger Documentation**
  Run: `make docs`
  Expected output: Swagger documentation successfully regenerated in `api/swagger/`.

- [ ] **Step 4: Verify full project compilation**
  Run: `make vet && make lint && make test`
  Expected output: Success

- [ ] **Step 5: Commit changes**
  Run: `git add internal/api/handler/post_handler.go cmd/main.go api/swagger/ && git commit -m "feat: refactor API handlers and routes for idempotent likes/unlikes"`
