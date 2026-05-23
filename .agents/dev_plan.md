# Development Plan: Post Comments Feature

This document outlines the design and implementation plan for the new Post Comments feature, following the Clean Architecture pattern used in this project.

## 1. Database Schema & Migration

A new migration will be created to add the `comments` table.

**Table:** `comments`
- `id`: bigint (Primary Key)
- `post_id`: bigint (Foreign Key referencing `posts.id`)
- `user_id`: bigint (Foreign Key referencing `users.id`)
- `content`: text
- `created_at`: timestamp
- `updated_at`: timestamp
- `deleted_at`: timestamp (for soft deletes)

*Decision:* Use `user_id` instead of `username` as the foreign key for better relational normalization and indexing.

## 2. Domain Model

**File:** `internal/core/domain/comment.go`

```go
package domain

import "time"

type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func (c *Comment) Validate() error {
	if len(c.Content) < 1 {
		return fmt.Errorf("%w: comment must be at least 1 character", ErrValidationFailed)
	}
	return nil
}
```

## 3. Architecture & Business Logic (Clean Architecture)

### 3.1. Ports
**File:** `internal/core/ports/comment.go`
- `CommentRepository`: Interface defining `Create(ctx context.Context, comment *domain.Comment) error`.
- `CommentUseCase`: Interface defining `AddComment(ctx context.Context, postID int, userID int, content string) error`.

### 3.2. UseCase
**File:** `internal/core/usecase/comment.go`
- Implements `CommentUseCase.AddComment`.
- **Workflow:**
  1. Validate the content by creating a domain `Comment` entity.
  2. Call `PostRepository.GetByID` to ensure the post exists. If not, return `domain.ErrPostNotFound`.
  3. Call `CommentRepository.Create` to persist the comment.

### 3.3. Adapter (Database)
**File:** `internal/adapters/postgres/comment_repository.go`
- Implements `CommentRepository` using GORM.
- Defines a GORM model for the `comments` table.

## 4. API Layer

**Endpoint:** `POST /posts/{id}/comments`
**File:** `internal/api/handlers/comment.go`

### Request Payload
```json
{
    "content": "Good post!"
}
```

### Handler Logic
1. Require authentication using the existing JWT middleware.
2. Extract the `user_id` from the JWT context.
3. Parse the `id` from the URL path for the `post_id`.
4. Bind and validate the JSON request body.
5. Call `CommentUseCase.AddComment`.

### Response Mapping
- **201 Created**: Empty body. (Success)
- **400 Bad Request**: `{"error": "Comment must be at least 1 character"}` (Validation failure)
- **401 Unauthorized**: `{"error": "Unauthorized"}` (Missing/invalid JWT)
- **404 Not Found**: `{"error": "Post not found"}` (Post does not exist)
- **500 Internal Server Error**: `{"error": "Something went wrong"}` (Unexpected errors)

## 5. Implementation Steps
1. Create SQL migration file for `comments` table and run it.
2. Create the `domain.Comment` entity.
3. Define the `ports` interfaces (`CommentRepository`, `CommentUseCase`).
4. Implement the `postgres.CommentRepository`.
5. Implement the `usecase.CommentUseCase` (requires injecting `PostRepository`).
6. Create HTTP request/response DTOs for the comment.
7. Implement `CommentHandler`.
8. Register the new route in the Echo router, protected by JWT middleware.
9. Write unit tests for the UseCase.