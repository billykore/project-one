# Implementation Plan: Separate Post Commands and Queries

**Branch**: `006-post-cqrs` | **Date**: 2026-09-17 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/006-post-cqrs/spec.md`

## Summary

Split the existing all-purpose post application boundary and post repository into focused command and query counterparts, then split the HTTP handler in the same way. Preserve every current route, DTO, error, authorization, pagination, feature-gating, idempotency, and notification behavior. This is a synchronous in-process CQRS refactor: it introduces neither a new read model nor asynchronous command processing.

## Technical Context

**Language/Version**: Go 1.26.2  
**Primary Dependencies**: Echo 4.15, GORM, GoMock, Testify, Swaggo  
**Storage**: Separate command and query repository adapters over the existing PostgreSQL data store; no schema or query-semantic changes  
**Testing**: Go unit and GORM adapter regression tests with regenerated GoMock ports; focused post, feed, and comment tests; baseline and post-change `make test-cover` comparison; reproducible post latency benchmark; `npm test -- --run` in `web/`; then `make check`  
**Target Platform**: Containerized Linux HTTP service  
**Project Type**: Full-stack web application; this change is backend-only  
**Performance Goals**: Preserve p95 read latency below 200 ms and write latency below 500 ms under normal load  
**Constraints**: No route, request/response, error, authorization, database schema, query behavior, feature-flag, notification, or frontend behavior changes; no new dependencies  
**Scale/Scope**: Eight post operations across nine existing routes, plus command-side comment validation, feed reads, generated mocks, and the application composition root

## Constitution Check

### Pre-design

- **Clean Architecture — PASS**: New driving and driven ports remain in `internal/core/ports`; use cases continue to depend only on ports; separate adapters remain in the adapter layer; handlers remain delivery adapters.
- **Testing — PASS**: Existing GoMock generation supports the split. Command/query post, feed, and comment tests will use their narrow ports; like-status query coverage is added.
- **User experience — N/A**: The frontend and external HTTP behavior are unchanged.
- **Performance — PASS**: The same repository calls and cursor pagination remain; no new query, projection, or unbounded read is introduced.
- **Security and authorization — PASS**: Existing authentication middleware, ownership checks, validation, and creation feature gate are retained.
- **Workflow — PASS**: Generated mocks, adapter and use-case tests, baseline coverage comparison, reproducible latency validation, frontend tests, Swagger verification, and `make check` remain required.

### Post-design

All gates remain satisfied. Command processing retains command-side aggregate reads required to validate a write, while query processing has no publisher, feature evaluator, user lookup, or mutating repository capability. No constitution exception is required.

## Project Structure

### Documentation (this feature)

```text
specs/006-post-cqrs/
├── plan.md
├── research.md
├── data-model.md
├── contracts/
│   └── post-http.md
├── coverage-baseline.txt                         # pre-refactor total coverage record
└── quickstart.md
```

### Source Code (repository root)

```text
cmd/main.go                                      # composition root and route registration
scripts/benchmark-post-cqrs.sh                   # reproducible post endpoint latency check
internal/core/ports/post.go                      # command and query driving ports
internal/adapters/repository/post_command_repository.go # command adapter over existing DB
internal/adapters/repository/post_query_repository.go   # query adapter over existing DB
internal/core/usecase/post_command_usecase.go    # create/update/delete/like/unlike
internal/core/usecase/post_query_usecase.go      # post/list/like-status reads
internal/core/usecase/comment_usecase.go         # command-side post precondition port
internal/core/usecase/feed_usecase.go            # query-side post repository port
internal/core/usecase/post_*_usecase_test.go     # focused command and query tests
internal/core/ports/mocks/mock_post.go           # generated command/query mocks
internal/api/handler/post_command_handler.go     # state-changing post routes
internal/api/handler/post_query_handler.go       # read-only post routes
internal/api/handler/user_handler.go             # query-only post dependency
```

**Structure Decision**: Replace the combined post use case, repository, and handler with command/query counterparts in the existing Clean Architecture layers. Command and query repository adapters are distinct concrete types over the same database and shared post model; a small private lookup helper preserves identical not-found and soft-delete behavior. Keep DTOs, routes, middleware, database schema, and frontend proxies unchanged. Keep comment creation on the command handler because it is an existing state-changing route that already delegates to the separate comment use case; moving it is unrelated to the post CQRS split.

## Complexity Tracking

No constitution violations or additional complexity are required.
