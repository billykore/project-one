## Summary
This PR implements the Post Like and Unlike toggle features, adding the core domain logic, Postgres repository, use case, and HTTP handlers. The implementation follows the Clean Architecture pattern established in the project.

## Critical Issues

### 1. `gorm.ErrDuplicatedKey` will never trigger
By default, GORM does not translate database-specific errors into GORM errors like `gorm.ErrDuplicatedKey` unless explicitly configured. Since `cmd/main.go` uses an empty `gorm.Config{}`, `errors.Is(err, gorm.ErrDuplicatedKey)` in `like_repository.go` will always be false. A duplicate like attempt will bubble up as a raw `*pgconn.PgError` and result in an unhandled 500 Internal Server Error.

**Fix**: Enable `TranslateError: true` in `cmd/main.go` inside `setupDB`:
```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    TranslateError: true,
})
```
*(Alternatively, check the PostgreSQL error code for unique violation (`23505`) directly).*

### 2. Unhandled Concurrent Toggles Cause 500 Errors
In `ToggleLike` (`like_usecase.go`), the logic checks `Exists` and then branches to `Create` or `Delete`. If a user double-clicks the like button (concurrent requests), both might read `Exists = false`, and both will attempt to `Create`. The second one will fail with `domain.ErrAlreadyLiked` (assuming the GORM issue above is fixed), which the use case currently logs and returns as `domain.ErrInternalServer`, resulting in a 500 response.

**Fix**: Handle these expected domain errors gracefully to ensure the endpoint behaves idempotently. For example:
```go
if err := u.likeRepo.Create(ctx, like); err != nil {
    if !errors.Is(err, domain.ErrAlreadyLiked) {
        u.log.Error(ctx, "failed to create like", "postID", postID, "error", err)
        return false, 0, domain.ErrInternalServer
    }
    // If it's already liked due to concurrent request, proceed to get count
}
```
*(Apply the same idempotent handling for `Delete` and `domain.ErrNotLiked`).*

## Suggestions for Improvement

### 1. Reduce Database Queries (Performance)
Currently, `ToggleLike` executes up to 4 sequential database queries per request (`GetByIDOnly`, `Exists`, `Create`/`Delete`, `CountByPostID`). You can eliminate the first two queries by relying entirely on your database constraints:
- Try to `Create` the like directly. 
- If it fails with `ErrDuplicatedKey`, you know it's already liked, so fallback to `Delete`.
- If it fails with a Postgres Foreign Key Violation, you know the `postID` does not exist, so you can safely return `ErrPostNotFound`.
This approach eliminates TOCTOU (Time-of-Check to Time-of-Use) race conditions completely and reduces DB load.

### 2. Denormalize Like Count (Performance)
`CountByPostID` uses a `SELECT COUNT(*)` on `post_likes`. As the application scales, this O(N) operation will become a bottleneck for popular posts. Consider adding a `like_count INT` column to the `posts` table and updating it atomically (e.g., `UPDATE posts SET like_count = like_count + 1 WHERE id = ?`) inside a transaction when a like is created or deleted.

### 3. Defensive Validation
While `username` is derived securely from the JWT, it’s defensive practice to add `if strings.TrimSpace(username) == ""` validation inside the UseCase along with the existing `postID <= 0` check.
