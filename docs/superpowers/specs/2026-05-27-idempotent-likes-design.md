# Spec: Idempotent Likes API Design

This specification defines the refactoring of the Post Like API to be idempotent.

## Goal
Currently, liking and unliking a post is handled via a single `POST /posts/:id/likes` endpoint which toggles the status. If a request is retried or sent twice due to user error or network fluctuation, the state alternates. This design refactors the API to expose two separate, idempotent operations:
- `POST /posts/:id/likes` to Like a post (always results in the post being liked).
- `DELETE /posts/:id/likes` to Unlike a post (always results in the post being unliked).

## Proposed Changes

### 1. Ports / Interfaces (`internal/core/ports/post.go`)
Refactor the `PostUseCase` interface to split the toggle operation:

```diff
- 	// ToggleLike toggles the like status for a given post ID and username.
- 	ToggleLike(ctx context.Context, postID int, username string) (liked bool, likeCount int, err error)
+ 	// LikePost likes a post by the given username. If already liked, it behaves idempotently.
+ 	LikePost(ctx context.Context, postID int, username string) (likeCount int, err error)
+ 	// UnlikePost unlikes a post by the given username. If not liked, it behaves idempotently.
+ 	UnlikePost(ctx context.Context, postID int, username string) (likeCount int, err error)
```

### 2. Business Logic / Use Cases (`internal/core/usecase/post_usecase.go`)
Remove `ToggleLike` and implement `LikePost` and `UnlikePost`:

- **LikePost**:
  - Checks if the like relationship already exists via `likeRepo.Exists`.
  - If it exists, does not create or increment the count (returns current count).
  - If it does not exist, creates the like via `likeRepo.Create` and increments the count via `postRepo.IncrementLikeCount`.
  
- **UnlikePost**:
  - Checks if the like relationship exists via `likeRepo.Exists`.
  - If it does not exist, does not delete or decrement (returns current count).
  - If it exists, deletes the like via `likeRepo.Delete` and decrements the count via `postRepo.IncrementLikeCount`.

### 3. API Handler (`internal/api/handler/post_handler.go`)
- Update `ToggleLike` method to `LikePost` (handling `POST /posts/:id/likes`).
- Create `UnlikePost` method (handling `DELETE /posts/:id/likes`).
- Update Swagger annotations for both endpoints.

### 4. Router Setup (`cmd/main.go`)
- Change the routing definition:
```go
posts.POST("/:id/likes", postHdl.LikePost)
posts.DELETE("/:id/likes", postHdl.UnlikePost)
```

## Verification & Testing
- Write unit tests in `internal/core/usecase/post_usecase_test.go` covering:
  - Liking an unliked post (success, count increases).
  - Liking an already liked post (idempotent success, count stays same).
  - Unliking a liked post (success, count decreases).
  - Unliking a post not liked (idempotent success, count stays same).
  - Validation failures and post not found scenarios.
- Run `go test ./...` to verify all tests pass.
