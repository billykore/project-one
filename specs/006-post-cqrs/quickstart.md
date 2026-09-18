# Quickstart: Validate Post CQRS Separation

**Date**: 2026-09-17 | **Feature**: `006-post-cqrs`

## Prerequisites

- Go 1.27.1 and the project dependencies available locally.
- For API smoke testing, use the backend setup from `README.md`; PostgreSQL and RabbitMQ are required only for full application runs.

## Validation steps

1. Regenerate mocks after changing the post ports:

   ```bash
   make mocks
   ```

2. Run focused post, feed, and comment use-case tests. They must demonstrate unchanged command behavior, command-side validation reads, and query paths—including like status and feed—without a mutating post repository:

   ```bash
   go test ./internal/core/usecase -run 'Test(Post(Command|Query)|FeedUseCase|CommentUseCase)' -count=1
   ```

3. Run the backend regression suite:

   ```bash
   make test
   ```

4. Verify generated API documentation remains in sync and run the required quality gate:

   ```bash
   make docs
   make check
   ```

5. With the backend running, replay the existing post API collection at `test/api/postman-collection.json`. Confirm the routes in [post-http.md](contracts/post-http.md) retain their established successful and error responses.

## Expected outcome

- The application composes separate command/query post repositories into matching use cases; the feed receives only the query repository and comment creation receives only the command repository.
- Reads do not create, update, delete, increment, decrement, or publish state.
- Existing clients observe no route, payload, status, pagination, ownership, feature-gate, idempotency, or notification behavior change.
