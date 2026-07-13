# Tasks: Global Error Handling Middleware

**Input**: Design documents from `/specs/002-global-error-handling/`

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅, quickstart.md ✅

**Tests**: Included — per Constitution §II, all use cases and adapters require tests. Middleware and registry are adapter-layer components.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Backend root**: Repository root (`internal/`, `cmd/`)
- **Middleware**: `internal/api/middleware/`
- **DTOs**: `internal/api/dto/`
- **Domain**: `internal/core/domain/`
- **Handlers**: `internal/api/handler/`
- **Entry point**: `cmd/main.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the foundational types and constants that all phases depend on

- [x] T001 Create error code constants (18 codes: 17 domain + 1 default) as Go string constants in `internal/core/domain/errors.go`
- [x] T002 [P] Create `ErrorMapping` struct and `ErrorRegistry` type with `Register()` method in `internal/api/middleware/error_registry.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core middleware infrastructure — MUST be complete before ANY user story implementation

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 Populate ErrorRegistry with all 17 domain sentinel → HTTP mapping entries (404/400/401/409/422/500) in `internal/api/middleware/error_registry.go`
- [x] T004 [P] Restructure `ErrorResponse` DTO to include `Code`, `Message`, `RequestID`, and `Details` fields in `internal/api/dto/error_dto.go`
- [x] T005 [P] Implement validator error parser that converts `validator.ValidationErrors` to `[]dto.ErrorDetail` in `internal/api/middleware/error_handler.go`
- [x] T006 Implement the custom `echo.HTTPErrorHandler` that maps errors via registry, formats JSON response, and sanitizes messages in `internal/api/middleware/error_handler.go`
- [x] T007 Implement the thin middleware wrapper (`ErrorMiddleware`) that intercepts non-nil errors from `next(c)`, looks up status via registry, and calls `c.Error()` to trigger HTTPErrorHandler in `internal/api/middleware/error_handler.go`
- [x] T008 Register `ErrorMiddleware` in Echo chain after `RequestID` and `Recover`, before route groups; set custom `e.HTTPErrorHandler` in `cmd/main.go`
- [x] T008a [P] Verify custom HTTP status override mechanism (FR-008) — detect `*echo.HTTPError` in middleware and pass through its Code without registry override; add test covering `echo.NewHTTPError(422, ...)` → 422 response in `internal/api/middleware/error_handler_test.go`
- [x] T009 [P] Unit tests for ErrorRegistry (mapping correctness, unknown error fallback, errors.Is chain walking, 17 sentinel coverage) in `internal/api/middleware/error_registry_test.go`
- [x] T010 [P] Unit tests for HTTPErrorHandler (status code mapping, JSON structure, request_id from context, Content-Type header, nil-error passthrough) in `internal/api/middleware/error_handler_test.go`
- [x] T011 [P] Unit tests for structured ErrorResponse DTO (JSON marshaling, omitempty on Details, field naming) in `internal/api/dto/error_dto_test.go`

**Checkpoint**: Foundation ready — middleware is registered and processes errors, but existing handlers still use manual `c.JSON()` and are unaffected (middleware dormant on nil returns). User story implementation can now begin.

---

## Phase 3: User Story 1 - Backend Developers Write Handlers Without Error-Mapping Boilerplate (Priority: P1) 🎯 MVP

**Goal**: Refactor all 6 handler files to return domain errors directly instead of calling `c.JSON(statusCode, dto.ErrorResponse{...})`, eliminating ~126 redundant error-handling blocks.

**Independent Test**: Run `go test ./...` — all existing tests must pass with semantically equivalent HTTP status codes. Call any endpoint with invalid input and verify the response follows the new structured error format from the middleware (not the old flat format).

### Implementation for User Story 1

- [x] T012 [US1] Refactor `user_handler.go` — replace all `c.JSON(status, dto.ErrorResponse{...})` error returns with `return domain.ErrXxx` (or `return err` for use-case errors); remove manual `errors.Is()` checks and `h.log.Error()` calls from handler methods in `internal/api/handler/user_handler.go`
- [x] T013 [P] [US1] Refactor `comment_handler.go` — replace all `c.JSON(status, dto.ErrorResponse{...})` error returns with `return err`; remove manual error mapping in `internal/api/handler/comment_handler.go`
- [x] T014 [P] [US1] Refactor `feed_handler.go` — replace all `c.JSON(status, dto.ErrorResponse{...})` error returns with `return err` in `internal/api/handler/feed_handler.go`
- [x] T015 [P] [US1] Refactor `notification_handler.go` — replace all `c.JSON(status, dto.ErrorResponse{...})` error returns with `return err` in `internal/api/handler/notification_handler.go`
- [x] T016 [P] [US1] Refactor `websocket_handler.go` — replace all `c.JSON(status, dto.ErrorResponse{...})` error returns with `return err` in `internal/api/handler/websocket_handler.go`
- [x] T017 [US1] Refactor `post_handler.go` — replace all `c.JSON(status, dto.ErrorResponse{...})` error returns with `return err`; add `log ports.Logger` field to `PostHandler` struct and constructor; wire logger in `cmd/main.go` in `internal/api/handler/post_handler.go`
- [x] T018 [US1] Refactor `authorization.go` middleware — replace `c.JSON(401, dto.ErrorResponse{...})` with `return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")` in `internal/api/middleware/authorization.go`
- [x] T019 [US1] Integration test: verify all 17 sentinel errors produce the correct HTTP status code documented in data-model.md mapping table; create test in `internal/api/middleware/error_handler_test.go`

**Checkpoint**: All handlers return errors directly. The middleware formats all error responses. `make test` passes. MVP is deployable — error responses follow the new structure, status codes are preserved, and boilerplate is eliminated.

---

## Phase 4: User Story 2 - API Consumers Receive Structured, Machine-Readable Error Responses (Priority: P2)

**Goal**: Validation errors include per-field `details` array. All error responses validate against the JSON schema contract. Frontend can build a single error-handling utility for all endpoints.

**Independent Test**: Call 3 endpoints producing different error types (validation failure, not-found, unauthorized) and verify all responses pass JSON Schema validation against `contracts/error-response.schema.json`.

### Implementation for User Story 2

- [x] T020 [US2] Implement field-level validation detail injection in the error handler — when the error wraps a `validator.ValidationErrors`, populate `ErrorResponse.Details` with per-field objects (`field`, `reason`, `message`) in `internal/api/middleware/error_handler.go`
- [x] T021 [US2] Ensure `Details` field is `omitempty` — non-validation errors produce no `details` key in JSON response in `internal/api/dto/error_dto.go`
- [ ] T022 [US2] Verify handlers that return domain-wrapped validation errors (e.g., `domain.User.Validate()` wrapping `ErrValidationFailed`) pass the validation details through; update handler code if needed in `internal/api/handler/user_handler.go`
- [ ] T023 [US2] Contract test: send requests to 3+ endpoints that produce validation, not-found, and unauthorized errors; validate each response body against `contracts/error-response.schema.json` in `internal/api/middleware/error_handler_test.go`
- [ ] T024 [US2] Verify error code stability — all 18 error codes (17 domain + 1 default) match the documented codes in data-model.md; add test in `internal/core/domain/errors_test.go`

**Checkpoint**: Structured error responses with field-level details. All responses conform to the JSON Schema contract. API consumers can rely on consistent format.

---

## Phase 5: User Story 3 - Operations Engineers Debug Issues with Request-Correlated Error Logs (Priority: P3)

**Goal**: Every error produces a single structured log entry with `request_id`, `method`, `path`, `status`, `error_code`, `error`, and `user` fields. Non-production includes stack trace. Production sanitizes response messages.

**Independent Test**: Trigger a 500 error, capture `request_id` from response, search server logs for that ID — verify exactly one structured log entry exists with all required fields.

### Implementation for User Story 3

- [x] T025 [US3] Implement structured error logging in the HTTPErrorHandler — log at Warn level for 4xx, Error level for 5xx, with fields: `request_id`, `method`, `path`, `status`, `error_code`, `error`, `user` in `internal/api/middleware/error_handler.go`
- [x] T026 [US3] Extract authenticated username from `c.Get("username")` for the `user` log field; default to `"anonymous"` when no auth context exists in `internal/api/middleware/error_handler.go`
- [x] T027 [US3] Implement conditional stack trace capture via `runtime/debug.Stack()` when `config.App.Env != "production"`; append `stack_trace` field to log entry in `internal/api/middleware/error_handler.go`
- [x] T028 [US3] Implement production sanitization — when `config.App.Env == "production"`, the error response `Message` field must be the registry-default message (never `err.Error()`); log the full error server-side in `internal/api/middleware/error_handler.go`
- [ ] T029 [US3] Integration test: trigger a 500 error in non-production mode, verify log output contains all 7 required fields plus `stack_trace` in `internal/api/middleware/error_handler_test.go`
- [ ] T030 [US3] Integration test: trigger a 500 error in production-like mode (set env), verify response body contains generic message (no raw Go error, no stack trace) and log entry lacks `stack_trace` in `internal/api/middleware/error_handler_test.go`

**Checkpoint**: All errors are logged with full context for debugging. Production responses are sanitized. Operations can trace errors end-to-end via request IDs.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, final validation, and cleanup

- [x] T031 [P] Update Swagger annotations for all handler endpoints — replace `@Failure` response examples with the new structured error format (`{"error": {"code": "...", "message": "...", "request_id": "..."}}`) in all handler files under `internal/api/handler/`
- [ ] T032 Run `make docs` to regenerate Swagger spec in `api/swagger/`
- [ ] T033 Run quickstart.md validation — execute all 8 scenarios and confirm expected outcomes
- [x] T034 Run full test suite `make test` and verify zero regressions (all ~200 existing tests pass)
- [x] T035 Run `make check` (docs + vet + lint + test) and verify all gates pass
- [ ] T036 [P] Add godoc comments to all new exported types (`ErrorCode`, `ErrorMapping`, `ErrorRegistry`, `ErrorResponse`, `ErrorDetail`) and functions (`NewErrorRegistry`, `ErrorMiddleware`, `NewErrorHandler`)
- [x] T037 [P] Verify SC-002 — run `grep -c 'c.JSON.*ErrorResponse'` across `internal/api/handler/*.go`; confirm fewer than 10 remaining instances (from original ~126) and document the count
- [ ] T038 Document the 3-step process for adding a new domain error (SC-003): (1) define sentinel in domain file, (2) add error code constant in `internal/core/domain/errors.go`, (3) register mapping in ErrorRegistry; add as godoc on `ErrorRegistry.Register()` and/or `internal/api/middleware/README.md`

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1: Setup
    └── Phase 2: Foundational (BLOCKS all user stories)
            ├── Phase 3: User Story 1 (P1) 🎯 MVP
            │       └── Phase 4: User Story 2 (P2)
            │               └── Phase 5: User Story 3 (P3)
            └── (US2 depends on US1 — handlers must return errors before details parsing works)
                (US3 depends on US1/US2 — logging happens on the error path built in US1)
```

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Phase 2 (Foundational). No dependency on US2 or US3.
- **User Story 2 (P2)**: Depends on US1 — validator errors must be returned from handlers (US1) before details injection can work.
- **User Story 3 (P3)**: Depends on US1/US2 — logging is layered onto the error path; sanitization depends on the response format (US2).

### Within Each Phase

| Phase | Execution Order |
|-------|----------------|
| Phase 1 | T001 → T002 (T002 depends on T001 for error code types) |
| Phase 2 | T003 depends on T002 → T004, T005 in parallel → T006 → T007 → T008 → T008a, T009, T010, T011 in parallel |
| Phase 3 | T012 → T013–T016 in parallel → T017 → T018 → T019 |
| Phase 4 | T020 → T021 → T022 → T023, T024 in parallel |
| Phase 5 | T025 → T026 → T027, T028 in parallel → T029, T030 in parallel |
| Phase 6 | T031, T036, T037, T038 all parallelizable after prior phases complete; T032 depends on T031; T033–T035 sequential |

### Parallel Opportunities

| Phase | Parallel Tasks |
|-------|---------------|
| Phase 1 | T002 can start alongside T001 (struct definition only, no import of constants needed if coded carefully) |
| Phase 2 | T004 ∥ T005 (different files); T008a ∥ T009 ∥ T010 ∥ T011 (different test files) |
| Phase 3 | T013 ∥ T014 ∥ T015 ∥ T016 (different handler files, no shared state) |
| Phase 4 | T023 ∥ T024 (different test files) |
| Phase 5 | T027 ∥ T028 (different concerns, same file but non-overlapping); T029 ∥ T030 (different test cases) |
| Phase 6 | T031 ∥ T036 ∥ T037 ∥ T038 (different files/documentation); T033–T035 sequential (validation steps) |

---

## Parallel Example: Phase 3 (User Story 1 — Handler Refactoring)

```bash
# After T012 (user_handler.go) is complete, refactor remaining handlers in parallel:
# Terminal 1:
#   Refactor comment_handler.go (T013)
# Terminal 2:
#   Refactor feed_handler.go (T014)
# Terminal 3:
#   Refactor notification_handler.go (T015)
# Terminal 4:
#   Refactor websocket_handler.go (T016)

# After all parallel refactors complete:
#   Refactor post_handler.go (T017) — depends on logger wiring in main.go
#   Refactor authorization.go (T018)
#   Run integration test (T019)
```

---

## Implementation Strategy

### MVP Scope (User Story 1 Only)

1. Complete Phase 1 (Setup) → Phase 2 (Foundational) — ~2 hours
2. Complete Phase 3 (User Story 1) — ~3 hours
3. **Deploy**: Handlers return errors, middleware formats responses, status codes preserved
4. **Validation**: `make test` passes, all endpoints return correct HTTP status codes

### Incremental Delivery

| Milestone | Phases | Value Delivered |
|-----------|--------|-----------------|
| **M1** | 1 + 2 + 3 | Error boilerplate eliminated; consistent error format; all status codes preserved |
| **M2** | +4 | Field-level validation details; frontend can build single error utility |
| **M3** | +5 | Request-correlated structured logging; production-safe error messages |
| **M4** | +6 | Polished API docs; validated against quickstart scenarios |

### Rollback Safety

Each handler refactoring is independent — if a specific handler causes issues, revert that single file without affecting others. The middleware coexists with old-style handlers because it's a no-op on nil returns.

---

## Phase 7: Convergence

**Purpose**: Cleanup and verification items identified during convergence analysis. One finding — dead code removal.

- [x] T039 Remove deprecated `dto.ErrorResponse` struct from `internal/api/dto/error_dto.go` after T031 (Swagger annotation updates) is complete and the struct has zero code references per `dead code` (unrequested)
