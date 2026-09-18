# Research: Post CQRS Split

**Date**: 2026-09-17 | **Feature**: `006-post-cqrs`

## Decision: Split the existing driving port by capability

Create two post-facing capabilities rather than adding a command bus, mediator, event store, read database, or projection.

- **Command capability**: create, update, delete, like, unlike.
- **Query capability**: retrieve one post, retrieve a paginated post list, retrieve like status.

**Rationale**: The current use case already contains these synchronous operations and their existing repositories expose the required reads and writes. A capability split makes callers depend only on what they use while preserving all behavior.

**Alternatives considered**:

- A mediator or command-bus package — rejected: it adds indirection without a second dispatch mechanism or cross-cutting command pipeline.
- Separate read storage or projections — rejected: no reporting/read-scale requirement exists and it would alter consistency behavior.
- Keeping one post interface with grouped methods — rejected: callers would still depend on unrelated operations, which fails the CQRS goal.

## Decision: Split post persistence by command/query capability

Replace the combined post persistence port with distinct command and query ports, implemented by distinct adapter types over the existing database and post model.

- **Command repository**: create, update, delete, adjust a like count, and load an aggregate when a command needs current state or ownership validation.
- **Query repository**: retrieve a public post, retrieve an author's paginated posts, and retrieve the feed.

The command port retains the existing owner-scoped and ID-only lookups as write preconditions. The query port exposes public post lookup, author listing, and feed retrieval. Each port is mocked independently; the same private adapter lookup preserves shared not-found handling.

**Rationale**: A command must be able to load the aggregate it will validate or mutate; injecting a query repository into a command would reintroduce cross-side coupling. The feed is a pure reader and must use the query repository.

**Alternatives considered**:

- One concrete repository returned as two interfaces — rejected: it does not meet the explicit concrete-repository split.
- A second database, projection, or replica — rejected: the specification explicitly retains the existing data store.
- Duplicate all lookup code — rejected: a small private lookup helper better preserves identical error and soft-delete behavior.

## Decision: Allocate only necessary dependencies

The command implementation keeps the command post repository, like repository, user repository, publisher, logger, and feature evaluator. The query implementation uses only the query post repository, like repository, and logger. The feed use case uses the query post repository; comment creation uses the command post repository for its write precondition.

**Rationale**: Creation needs feature evaluation; like notifications need user lookup and publishing; command-side preconditions need current aggregate state; reads need neither mutation capability nor publishing. This makes it impossible for query code to publish a notification or use a mutating post repository accidentally.

**Alternatives considered**: Sharing a common service with all dependencies — rejected: it preserves the coupling CQRS is intended to remove.

## Decision: Split handlers without changing routes

Use a command handler for authenticated state-changing post routes and a query handler for post reads. The post-detail query may continue to obtain comment data through the existing comment use case because it is still a read-only response composition. Keep comment creation on the command handler; it already delegates to the dedicated comment use case and moving it would be unrelated churn.

**Rationale**: Route ownership follows request behavior while retaining all existing middleware, Swagger annotations, DTO mapping, and status/error handling.

**Alternatives considered**: Leave one handler with both dependencies — rejected: it leaves the delivery layer coupled to both sides. Move comment creation to a new handler — rejected: no comment CQRS scope exists.

## Decision: Regenerate existing mocks and split unit tests

Regenerate `mock_post.go` from the existing ports source after the driving and repository interface splits. Move existing post tests to command/query-focused files or sections, update feed and comment tests to their narrow repository mocks, and add like-status query coverage, which the current use-case suite lacks.

**Rationale**: GoMock and the `make mocks` target already support this change. Narrow mocks prove handler and user-profile callers cannot invoke unrelated operations.

**Alternatives considered**: Hand-maintain mocks or add a new mocking dependency — rejected: generated mocks are already committed and supported by the project.

## Preserved-risk checklist

- Retain the creation feature gate in middleware and the use case.
- Retain ownership checks, command-side aggregate loading, error mapping, partial-update trimming, and cursor `limit + 1` semantics.
- Retain repository not-found mapping and soft-delete filtering in both command and query adapters.
- Retain idempotent like/unlike behavior and notification only when liking another user's post.
- Do not change the existing non-transactional like/count behavior; transaction or outbox work is out of scope.
- Preserve all HTTP paths, methods, authentication, DTO fields, status codes, and RFC 9457 errors.
