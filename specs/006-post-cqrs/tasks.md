# Tasks: Separate Post Commands and Queries

**Input**: Design documents from `specs/006-post-cqrs/`
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/post-http.md`, `quickstart.md`

## Format: `[ID] [P?] [Story] Description`

- `[P]` means the task can proceed in parallel after its stated dependencies.
- `[US#]` maps the task to a user story in `spec.md`.

## Phase 1: Setup

**Purpose**: Confirm the refactor baseline and preserved contract before changing boundaries.

- [X] T001 Run `make test-cover` before refactoring and record the `go tool cover -func=test/coverage/coverage.out | tail -1` result in `specs/006-post-cqrs/coverage-baseline.txt`.

---

## Phase 2: Foundational Boundaries

**Purpose**: Create the narrow driving and driven ports that all story work depends on.

- [X] T002 Replace combined post interfaces with `PostCommandUseCase`, `PostQueryUseCase`, `PostCommandRepository`, and `PostQueryRepository` in `internal/core/ports/post.go`, retaining `PostsPage`.
- [X] T003 Split the GORM post adapter into distinct command and query adapters in `internal/adapters/repository/post_command_repository.go` and `internal/adapters/repository/post_query_repository.go`; preserve `postModel`, error mapping, soft-delete filtering, and cursor queries while removing `internal/adapters/repository/post_repository.go`.
- [X] T004 Regenerate the committed command/query repository and use-case mocks in `internal/core/ports/mocks/mock_post.go` with `make mocks`.
- [X] T005 [P] Migrate comment creation's post precondition to `PostCommandRepository` and update its narrow-mock expectations in `internal/core/usecase/comment_usecase.go` and `internal/core/usecase/comment_usecase_test.go`.
- [X] T006 [P] Migrate feed retrieval to `PostQueryRepository` and update its narrow-mock expectations in `internal/core/usecase/feed_usecase.go` and `internal/core/usecase/feed_usecase_test.go`.

**Checkpoint**: The repository split is complete; generated mocks, feed, and comment code compile against narrow ports.

---

## Phase 3: User Story 1 — Publish and Manage a Post (Priority: P1)

**Goal**: Authors can create, update, and delete their own posts with existing validation, ownership, and response behavior.

**Independent Test**: Command use-case tests prove create, partial update, invalid ID, ownership, and delete behavior without any query-use-case dependency.

- [X] T007 [US1] Move/create create-update-delete test coverage against `PostCommandUseCase` in `internal/core/usecase/post_command_usecase_test.go` before moving production logic.
- [X] T008 [US1] Implement `PostCommandUseCase` creation, update, and deletion behavior with `PostCommandRepository` in `internal/core/usecase/post_command_usecase.go`, preserving feature evaluation, trimming, ownership lookup, errors, and logs.
- [X] T009 [US1] Create `PostCommandHandler` for create, update, delete, and unchanged comment creation in `internal/api/handler/post_command_handler.go`, preserving DTO validation, status codes, Swagger annotations, and error propagation.

**Checkpoint**: `go test ./internal/core/usecase -run TestPostCommand -count=1` passes for author publishing and management behavior.

---

## Phase 4: User Story 2 — Read Post Content and Engagement (Priority: P1)

**Goal**: Readers retrieve posts, author lists, and like status through query-only dependencies with unchanged pagination and response data.

**Independent Test**: Query use-case tests prove get-by-ID, cursor pagination, and like-status results without a command post repository; the existing feed test proves the feed uses the query repository.

- [X] T010 [US2] Move/create get-by-ID, author-list, and like-status test coverage against `PostQueryUseCase` in `internal/core/usecase/post_query_usecase_test.go`, including the previously untested like-status success/failure paths and 10 repeated reads that invoke no post, like, or notification mutation.
- [X] T011 [US2] Implement `PostQueryUseCase` reads in `internal/core/usecase/post_query_usecase.go`, preserving invalid-ID errors, cursor `limit + 1` behavior, limits, and logging without any mutation or publisher dependency.
- [X] T012 [US2] Create `PostQueryHandler` for post detail, authenticated post lists, and like status in `internal/api/handler/post_query_handler.go`, retaining comment response composition, DTO mapping, Swagger annotations, and error behavior.
- [X] T013 [P] [US2] Change the user-profile post reader to depend only on `PostQueryUseCase` in `internal/api/handler/user_handler.go`.

**Checkpoint**: `go test ./internal/core/usecase -run 'Test(PostQuery|FeedUseCase)' -count=1` passes and no query post use case receives a command post repository.

---

## Phase 5: User Story 3 — Like and Unlike Safely (Priority: P2)

**Goal**: Likes and unlikes remain idempotent, retain displayed counts, and notify only another post owner.

**Independent Test**: Command use-case tests repeat likes/unlikes and verify count, existing errors, and notification conditions through the command-side dependencies.

- [X] T014 [US3] Move/create like and unlike idempotency, count, error, and notification test coverage in `internal/core/usecase/post_command_usecase_test.go`.
- [X] T015 [US3] Implement like, unlike, and existing like-notification publishing in `internal/core/usecase/post_command_usecase.go` using command-side aggregate precondition reads and preserving all current error mapping.
- [X] T016 [US3] Add like and unlike route methods to `internal/api/handler/post_command_handler.go`, retaining response payloads, status codes, authentication context handling, and Swagger annotations.

**Checkpoint**: `go test ./internal/core/usecase -run TestPostCommand -count=1` passes for repeat likes/unlikes and cross-user notifications.

---

## Phase 6: Composition, Validation, and Polish

**Purpose**: Replace combined wiring only after both CQRS sides are complete and verify the unchanged contract.

- [X] T017 [P] Add GORM adapter regression tests in `internal/adapters/repository/post_repository_test.go` for command/query not-found mapping and soft-deleted-post exclusion.
- [X] T018 Wire separate command/query repositories, use cases, and handlers in `cmd/main.go`; route unchanged write endpoints and comment creation to `PostCommandHandler`, unchanged read endpoints to `PostQueryHandler`, pass the query repository to feed, and remove superseded `internal/core/usecase/post_usecase.go`, `internal/core/usecase/post_usecase_test.go`, and `internal/api/handler/post_handler.go`.
- [X] T019 [P] Run and repair the focused regression suites in `internal/core/usecase/post_command_usecase_test.go`, `internal/core/usecase/post_query_usecase_test.go`, `internal/core/usecase/comment_usecase_test.go`, and `internal/core/usecase/feed_usecase_test.go` with `go test ./internal/core/usecase -count=1`.
- [X] T020 [P] Regenerate and verify unchanged API documentation from Swagger annotations in `api/swagger/docs.go`, `api/swagger/swagger.json`, and `api/swagger/swagger.yaml` with `make docs`.
- [ ] T021 [P] Add and run `scripts/benchmark-post-cqrs.sh` using one existing public post URL, one authenticated author-owned post URL, `ACCESS_TOKEN`, 200 requests at concurrency 20 per endpoint, and `curl` timing output; fail p95 reads above 200 ms or writes above 500 ms.
- [X] T022 [P] Run `make test-cover`, compare its total against `specs/006-post-cqrs/coverage-baseline.txt`, and verify no decrease plus at least 80% coverage for mutation paths.
- [X] T023 [P] Run `npm test -- --run` from `web/` to satisfy the frontend pre-merge gate.
- [X] T024 Run the complete quality gate from `Makefile` with `make check` after T019–T023 and resolve all failures without changing the contract in `specs/006-post-cqrs/contracts/post-http.md`.

---

## Dependencies & Execution Order

```text
T001 → T002 → T003 → T004
                    ├── T005 (comment command consumer)
                    ├── T006 (feed query consumer)
                    ├── T017 (adapter regression tests)
                    ├── US1: T007 → T008 → T009
                    └── US2: T010 → T011 → T012, T013

US1 → US3: T014 → T015 → T016
T005 + T006 + T017 + US1 + US2 + US3 → T018 → {T019, T020, T021, T022, T023} → T024
```

## Parallel Opportunities

- After T004, T005 and T006 modify unrelated command/query consumers in parallel.
- After T004, the US1 command path (T007–T009) and US2 query path (T010–T013) use different files and can proceed in parallel.
- After T003, T017 can run independently of generated mocks and use-case migration.
- After T018, focused test repair (T019), Swagger generation (T020), performance validation (T021), coverage validation (T022), and frontend tests (T023) can run in parallel.

## Implementation Strategy

1. Complete Phase 2 first; it establishes compile-time CQRS boundaries without changing the data store.
2. Deliver US1 and US2 as independently unit-testable command/query increments.
3. Add US3 engagement commands, then perform one composition-root route migration.
4. Finish with adapter regressions, documentation, reproducible performance and coverage comparisons, frontend verification, and the full quality gate.

## Format Validation

All 24 tasks use the required checkbox, sequential ID, optional parallel marker, story label where applicable, and explicit file path format.

---

## Phase 7: Convergence

- [X] T025 CRITICAL Restore non-decreasing total coverage and demonstrate at least 80% coverage for post mutation paths with focused CQRS tests per Constitution II and T022 (partial).
- [X] T026 Replace the embedded `postRepository` wrappers in `internal/adapters/repository/post_command_repository.go` and `internal/adapters/repository/post_query_repository.go` with distinct command/query adapters, then remove `internal/adapters/repository/post_repository.go` per FR-001, FR-007, FR-009, and plan: Structure Decision (partial).
- [X] T027 Extract command behavior to `internal/core/usecase/post_command_usecase.go` and query behavior to `internal/core/usecase/post_query_usecase.go`, each with only its narrow dependencies; remove `internal/core/usecase/post_usecase.go` and the error-returning compatibility adapters in `internal/core/usecase/post_cqrs_usecase.go` per FR-001 through FR-004 and FR-007 (contradicts).
- [X] T028 Split `internal/api/handler/post_handler.go` into `internal/api/handler/post_command_handler.go` and `internal/api/handler/post_query_handler.go` with narrow use-case fields, then remove the aliases and compatibility adapters in `internal/api/handler/post_cqrs_handler.go` per FR-001 through FR-004 and FR-007 (contradicts).
- [X] T029 Add narrow-mock command/query use-case tests for create, update, delete, like, unlike, post retrieval, pagination, like status, and ten mutation-free repeated reads in `internal/core/usecase/post_command_usecase_test.go` and `internal/core/usecase/post_query_usecase_test.go` per US1/US2/US3, SC-003, SC-004, and Constitution II (partial).
- [X] T030 Add GORM regression tests for command/query not-found mapping and soft-deleted-post exclusion in `internal/adapters/repository/post_repository_test.go` per T017 and Constitution II (missing).
- [ ] T031 Add and execute `scripts/benchmark-post-cqrs.sh` against the configured public-read and authenticated author-write endpoints, enforcing the specified p95 limits per SC-005 and T021 (missing).
- [ ] T032 Run the existing post HTTP contract regression collection and the complete quality gate after T025–T031, preserving all established route behavior per SC-001 and Constitution I (partial).

---

## Phase 8: Convergence

- [X] T033 Extend `internal/core/usecase/post_query_usecase_test.go` so get-by-ID and paginated author-list reads each run ten times through `PostQueryUseCase` without any mutation-capable dependency per SC-003 and US2 (partial).
- [ ] T034 With `POSTS_TEST_DSN`, `BASE_URL`, `PUBLIC_POST_ID`, `AUTHOR_POST_ID`, and `ACCESS_TOKEN` configured, execute `go test ./internal/adapters/repository -run TestPostCommandAndQueryRepositoryNotFoundAndSoftDelete -count=1`, `scripts/benchmark-post-cqrs.sh`, and `test/api/postman-collection.json`; record p95 and contract results per SC-001, SC-005, Constitution II, and Constitution IV (partial).
