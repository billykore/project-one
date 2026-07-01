# Feeds API Design Spec

**Date:** 2026-07-01
**Status:** Approved

## Overview

Create a new `GET /feeds` API endpoint that returns a paginated feed of posts from users the authenticated user follows, plus their own posts. Uses cursor-based pagination for performance at scale.

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Pagination | Cursor-based (`created_at DESC, id DESC`) | Stable ordering, handles large datasets better than offset |
| Response shape | Envelope with `data`, `next_cursor`, `has_more` | Client needs cursor to paginate; counts are unreliable for cursor pagination |
| `message` field | Dropped | No equivalent in existing `Post` domain model |
| Nested comments | None | Client fetches comments separately via `GET /posts/:id/comments`; avoids expensive N+1 |
| Query approach | WHERE IN (followed usernames + self) | Simplest; composite index makes it fast; UNION approach reserved for scale-up |
| Cursor encoding | Base64-encoded JSON of `{created_at, id}` | Opaque, tamper-resistant, extensible |

## Architecture

```
Client → Echo Handler → FeedUseCase → FeedRepository (port) → GORM Adapter → PostgreSQL
```

Follows existing Clean Architecture layers. No new domain entity — reuses `domain.Post`.

## Files

| Layer | File | Purpose |
|-------|------|---------|
| Ports | `internal/core/ports/feed.go` | `FeedRepository` + `FeedUseCase` interfaces |
| UseCase | `internal/core/usecase/feed_usecase.go` | `FeedUseCase` implementation |
| Adapter | `internal/adapters/repository/feed_repository.go` | GORM `FeedRepository` |
| API DTO | `internal/api/dto/feed_dto.go` | `FeedResponse` envelope + cursor helpers |
| API Handler | `internal/api/handler/feed_handler.go` | `HandleGetFeed` |
| Routes | `cmd/main.go` | Register `GET /feeds` (auth required) |
| Migration | `db/migrations/000014_add_feed_index.up.sql` | Composite partial index |
| Migration | `db/migrations/000014_add_feed_index.down.sql` | Drop index |

## API Specification

### Request

```
GET /feeds?cursor=<opaque_base64>&limit=10
Authorization: Bearer <token>
```

| Param | Type | Required | Default | Max |
|-------|------|----------|---------|-----|
| `cursor` | string | No | (none — first page) | — |
| `limit` | int | No | 10 | 50 |

### Response (200 OK)

```json
{
  "data": [
    {
      "id": 42,
      "username": "alice",
      "title": "Hello World",
      "content": "Post content here",
      "tags": ["go", "api"],
      "like_count": 5,
      "created_at": "2026-07-01T10:00:00Z",
      "updated_at": "2026-07-01T10:00:00Z"
    }
  ],
  "next_cursor": "eyJjcmVhdGVkX2F0IjoiMjAyNi0wNy0wMVQxMDowMDowMFoiLCJpZCI6NDJ9",
  "has_more": true
}
```

- `next_cursor`: null when `has_more` is false
- `has_more`: true when there are more results beyond this page

### Error Responses

| Status | Error | Condition |
|--------|-------|-----------|
| 400 | `{"error":"invalid cursor"}` | Malformed cursor |
| 400 | `{"error":"limit must be between 1 and 50"}` | Invalid limit |
| 401 | `{"error":"unauthorized"}` | Missing/invalid token |
| 500 | `{"error":"internal server error"}` | DB or unexpected errors |

## Data Flow

1. Handler extracts `username` from auth context
2. Handler decodes cursor (if present), clamps limit to [1, 50]
3. `FeedUseCase.GetFeed(ctx, username, cursor, limit)`:
   - Resolve user via `UserRepository.GetUserByUsername`
   - Fetch followed usernames via `FollowRepository.GetFollowing` (large limit)
   - Build `usernames = [self] + followed_usernames`
   - Call `FeedRepository.GetFeed(ctx, usernames, cursor, limit+1)` (fetch one extra for `has_more`)
   - Build `FeedResult` with posts, encoded next cursor, has_more flag
4. Handler maps to `FeedResponse` DTO and returns 200

## SQL Query

```sql
SELECT * FROM posts
WHERE username IN (?, ?, ...)
  AND deleted_at IS NULL
  AND (created_at, id) < (?, ?)  -- cursor; omitted on first page
ORDER BY created_at DESC, id DESC
LIMIT ? + 1
```

## Index

```sql
CREATE INDEX idx_posts_username_created_at_id
ON posts (username, created_at DESC, id DESC)
WHERE deleted_at IS NULL;
```

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Empty feed (no follows, no posts) | `{"data":[], "has_more":false, "next_cursor":null}` |
| Cursor points past end | Empty data, has_more=false |
| Deleted post at cursor position | Skipped via `deleted_at IS NULL`; next valid posts returned |
| Concurrent new posts during pagination | Stable — cursor pins to a point in time; new posts appear on fresh first page |
| User follows themselves | Prevented by `ErrCannotFollowSelf`; not in followed list |
| Malformed cursor | 400 error |
| All posts deleted after cursor | Empty page, has_more=false |

## Test Plan

### Unit Tests (UseCase)
- GetFeed with valid user, follows, and posts → returns correct posts in order
- GetFeed with no follows → returns own posts only
- GetFeed with cursor → returns next page correctly
- GetFeed last page → has_more=false, next_cursor=null
- GetFeed user not found → error
- GetFeed with limit clamping → clamps to [1, 50]

### Integration Tests (Handler)
- GET /feeds without auth → 401
- GET /feeds with valid auth → 200 with envelope
- GET /feeds with invalid cursor → 400
- GET /feeds with limit=0 → 400
- GET /feeds with limit=100 → clamped to 50, returns 200

### Repository Tests
- GetFeed with multiple usernames → returns posts sorted correctly
- GetFeed cursor pagination → respects cursor boundary
- GetFeed has_more detection → returns limit+1 rows, correctly flags has_more
