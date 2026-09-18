# Feature Specification: Separate Post Commands and Queries

**Feature Branch**: `006-post-cqrs`  
**Created**: 2026-09-16  
**Status**: Draft  
**Input**: User description: "Refactor post_usecase.go to implement the CQRS pattern"

## Clarifications

### Session 2026-09-17

- Q: How should post persistence be separated? → A: Use distinct command and query post repositories over the existing data store; separate physical read storage is out of scope.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Publish and manage a post (Priority: P1)

An authenticated author can create, edit, or delete a post with the same validation, ownership protection, and visible results as before the refactor.

**Why this priority**: Publishing and managing content is the core post workflow; a refactor must not interrupt it.

**Independent Test**: An author creates, updates, and deletes a post, including attempts to modify a post they do not own, and receives the established outcomes.

**Acceptance Scenarios**:

1. **Given** an authenticated author who is allowed to publish, **When** they create a valid post, **Then** the post is stored and returned with the established successful result.
2. **Given** a post owned by the authenticated author, **When** they update or delete it, **Then** only that post changes or is removed and the established result is returned.
3. **Given** a post owned by another author, **When** an authenticated user tries to update or delete it, **Then** the action is rejected without changing the post.

---

### User Story 2 - Read post content and engagement (Priority: P1)

A reader can retrieve an individual post, an author's paginated post list, and their own like status without those requests changing post or engagement data.

**Why this priority**: Browsing is the primary consumption path and must remain safe, stable, and independent of write processing.

**Independent Test**: A reader retrieves a post, pages through an author's posts, and checks like status; repeated reads return the same data unless a separate write occurs.

**Acceptance Scenarios**:

1. **Given** an existing post, **When** a reader retrieves it or its like status, **Then** the returned content and engagement information match the stored state and no state changes.
2. **Given** an author with more posts than one page permits, **When** a reader follows the returned cursor, **Then** they receive the next page with the established page size and continuation behavior.

---

### User Story 3 - Like and unlike safely (Priority: P2)

An authenticated user can like or unlike a post without duplicate requests corrupting the displayed count, and the post owner receives the existing notification behavior.

**Why this priority**: Engagement affects both visible counts and notifications, so it is the most important write-side edge of the refactor after publishing.

**Independent Test**: A user likes and unlikes another user's post repeatedly and verifies the count, result, and notification behavior after each request.

**Acceptance Scenarios**:

1. **Given** an unliked post owned by another user, **When** a user likes it, **Then** the count increases once and the owner receives the established like notification.
2. **Given** a post already liked by the same user, **When** they like it again, **Then** the count is unchanged and the established idempotent result is returned.
3. **Given** a post not liked by the same user, **When** they unlike it, **Then** the count is unchanged and the established idempotent result is returned.

### Edge Cases

- Invalid or missing post identifiers are rejected without reading or changing post data.
- A request for a post that no longer exists returns the established not-found result and does not change engagement data.
- A disabled or unavailable post-creation entitlement prevents creation without storing a partial post.
- Repository, notification, or lookup failures preserve the established client-facing error behavior and do not report a successful command.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST separate post actions and post persistence access that change state from post requests and persistence access that only retrieve state.
- **FR-002**: The system MUST provide one focused command operation for creating, updating, deleting, liking, and unliking posts.
- **FR-003**: The system MUST provide one focused query operation for retrieving a post, retrieving a cursor-paginated author post list, and retrieving a user's like status.
- **FR-004**: Query operations MUST NOT create, update, delete, increment, decrement, or publish post-related state.
- **FR-005**: All existing post client interactions MUST retain their request data, response data, success status, error status, validation rules, ownership rules, and pagination behavior.
- **FR-006**: Commands MUST retain existing business behavior for creation eligibility, partial post updates, idempotent likes and unlikes, like counts, and notifications for likes of another user's post.
- **FR-007**: A caller needing a post command MUST NOT depend on unrelated post queries, and a caller needing a post query MUST NOT depend on unrelated post commands.
- **FR-008**: The system MUST continue to surface failures through the established error categories and MUST NOT expose internal processing details to clients.
- **FR-009**: The system MUST provide distinct post repository capabilities for commands and queries; command repository capabilities MAY change post state, while query repository capabilities MUST be read-only.

### Key Entities

- **Post**: Published content owned by an author, including its title, content, tags, creation details, and visible like count.
- **Like**: A user's engagement with one post, used to determine per-user like status and the post's aggregate count.
- **Post command**: A request that may change post or engagement state.
- **Post query**: A request that returns post or engagement state without changing it.
- **Post command repository**: The persistence capability used only by post commands to create, update, delete, or adjust post state.
- **Post query repository**: The persistence capability used only by post queries to retrieve posts and paginated post lists without changing state.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of established post endpoint regression scenarios retain the same externally observable request and response behavior after the refactor.
- **SC-002**: Each of the eight existing post actions is classified as exactly one command or one query, with no operation performing both roles.
- **SC-003**: Repeating each supported read request 10 times without an intervening write produces no change to post, like, or notification state.
- **SC-004**: 100% of tested duplicate like and unlike requests preserve the pre-existing idempotent count and outcome.
- **SC-005**: Post read and write requests continue to meet the project's existing user-visible response-time targets under normal load.
- **SC-006**: 100% of calls to post persistence use the corresponding command or query repository capability.

## Assumptions

- The refactor is internal: existing routes, request and response shapes, authentication, and authorization behavior remain unchanged.
- The current eight post actions are in scope: create, retrieve by ID, retrieve an author's posts, update, delete, like, unlike, and retrieve like status.
- Separate command and query post repositories use the existing data store and records; new physical read storage, replication, and projections are out of scope.
- Existing feature-entitlement, logging, and notification capabilities are retained; redesigning their behavior is out of scope.
- Commands and queries may share domain concepts and error rules, but callers receive only the capability needed for their operation.
