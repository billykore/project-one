### Summary
This PR successfully introduces the `GET /users/me/followers` endpoint, perfectly complementing the existing following API. The implementation efficiently determines mutual follower status using a single query with a `LEFT JOIN`, preventing the dreaded N+1 query problem. The code is clean, adheres strictly to the established architectural patterns, and includes appropriate default bounds for pagination and thorough Swagger documentation.

### Critical Issues
None identified. The code correctly handles edge cases, effectively guards against `nil` slice serialization by appropriately initializing the response array, and handles validation cleanly.

### Suggestions for Improvement

#### 1. HIGH: Non-Deterministic Pagination
**Issue:** In the repository query, ordering solely by `follows.created_at DESC` can lead to non-deterministic pagination. If multiple users happen to follow the current user at the exact same timestamp (which is entirely possible in high-concurrency environments or batch operations), the database does not guarantee a consistent row order for those tied records. This can cause followers to be skipped or duplicated across pages when the `offset` advances.
**Location:** `internal/adapters/repository/follow_repository.go`

```go
		Joins("LEFT JOIN follows AS mutual ON mutual.follower_id = follows.followed_id AND mutual.followed_id = follows.follower_id").
		Where("follows.followed_id = ?", followedID).
-		Order("follows.created_at DESC").
+		Order("follows.created_at DESC, follows.follower_id DESC").
		Limit(limit).Offset(offset).
```
**Step-by-step plan:** 
1. Open `internal/adapters/repository/follow_repository.go`.
2. Update the `Order` clause in `GetFollowers` to include a unique secondary sort key (like `follows.follower_id DESC`) to guarantee a stable, deterministic sort order. 
3. *Note: You should also apply this same fix to the `GetFollowing` method for consistency.*

#### 2. HIGH: Performance Bottleneck (Missing Index)
**Issue:** The new query heavily relies on the `WHERE follows.followed_id = ?` filter. Since the `follows` table's primary key is the composite `(follower_id, followed_id)`, PostgreSQL cannot efficiently utilize this B-tree index when filtering *only* by the second column. As a result, this query will trigger a sequential scan of the `follows` table, leading to an O(N) performance degradation as the user base scales.
**Location:** Database Migrations / `internal/adapters/repository/follow_repository.go`

**Step-by-step plan:** 
1. Generate a new database migration pair (e.g., `000005_add_followed_id_index_to_follows`).
2. In the `.up.sql` file, add a secondary index to support reverse lookups:
   ```sql
   CREATE INDEX idx_follows_followed_id ON follows(followed_id);
   ```
3. In the `.down.sql` file, add the teardown statement:
   ```sql
   DROP INDEX IF EXISTS idx_follows_followed_id;
   ```

#### 3. LOW: DTO Mapping Refactoring (Maintainability)
**Issue:** The iterative mapping logic in the `GetFollowers` handler is completely manual. While acceptable for a few fields, doing this inline inside the HTTP handler reduces the controller's readability and makes it slightly harder to test the mapping logic in isolation.
**Location:** `internal/api/handler/user_handler.go`

```go
	res := make([]dto.FollowerResponse, 0, len(followers))
	for _, f := range followers {
		res = append(res, dto.FollowerResponse{
			ID:         f.ID,
// ...
```
**Step-by-step plan:**
1. Consider extracting the `domain.Follower` to `dto.FollowerResponse` translation into a dedicated private mapping function within the handler file (e.g., `func toFollowerResponse(f domain.Follower) dto.FollowerResponse`).
2. Apply the mapper function inside the `for` loop to keep the main handler body lean and readable.