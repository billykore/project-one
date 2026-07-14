# Implementation Plan: RFC 9457 Problem Details Error Format

**Branch**: `003-rfc9457-error-format` | **Date**: 2026-07-14 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-rfc9457-error-format/spec.md`

## Summary

Migrate the existing global error handling middleware's HTTP response body from the custom `{"error": {"code": "...", "message": "...", ...}}` format to the RFC 9457 (STD 97) Problem Details for HTTP APIs standard. The middleware architecture (Echo HTTPErrorHandler, sentinel error mapping, structured logging) remains intact — only the serialization format and `Content-Type` header change. The frontend error-handling utilities are updated in tandem to parse the new `application/problem+json` responses.

## Technical Context

**Language/Version**: Go 1.26+ (backend), TypeScript 5.x / React 19 (frontend)

**Primary Dependencies**: Echo v4 (HTTP framework), GORM (ORM), Zerolog (logging), Validator v10 (validation), Swaggo (API docs), Next.js 16 (frontend framework), Tailwind CSS 4 (styling)

**Storage**: PostgreSQL via GORM (no schema changes; error format is transport-only)

**Testing**: GoMock + Testify (backend), Vitest (frontend)

**Target Platform**: Linux server (backend), Web browser (frontend)

**Project Type**: Web application (Go backend + Next.js frontend)

**Performance Goals**: Error responses are infrequent (<1% of requests). No performance regression: error handler must not add >1ms overhead vs current implementation.

**Constraints**: p95 <200ms read, <500ms write endpoints. Error handler must not allocate per-request dynamic memory that could be avoided (reuse type URI strings from static mapping).

**Scale/Scope**: ~18 domain sentinel errors, ~15 API endpoints, ~3 frontend error-handling files. Estimated ~12 files changed across backend and frontend.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Clean Architecture** | ✅ PASS | Error handler resides in `internal/api/middleware/` (API layer), DTOs in `internal/api/dto/` (API layer), sentinel errors in `internal/core/domain/` (Domain layer). No layer boundary violations. RFC 9457 format is a presentation concern confined to the API layer. |
| **II. Testing Standards** | ✅ PASS | Existing tests in `error_handler_test.go`, `error_registry_test.go`, and `error_dto_test.go` will be updated to validate RFC 9457 structure. Frontend tests in `web/tests/` will be updated. No coverage decrease. |
| **III. UX Consistency** | ✅ PASS | Frontend error display (modals, toasts, form field errors) preserves existing UX patterns. Only the data extraction path changes — the visual presentation is unaffected. |
| **IV. Performance** | ✅ PASS | Error handler is on the error path (infrequent). The static `type` URI and `title` strings are stored in the existing `errorMappings` map — no per-request allocation beyond the JSON serialization (which already happens today). |
| **Security** | ✅ PASS | `detail` field continues to be sanitized in production (mapped message, never raw error string). No sensitive data exposure changes. |
| **Code Quality** | ✅ PASS | No commented-out code, no dead imports. All changes pass `make check` (docs + vet + lint + test). |

**Gate result**: All principles pass. No violations to justify.

## Post-Design Constitution Re-Check

*Re-evaluated after Phase 1 design (data-model.md, contracts/, quickstart.md).*

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Clean Architecture** | ✅ PASS | Data model changes confined to `internal/api/dto/` and `internal/api/middleware/` (API layer). Domain sentinel errors in `internal/core/domain/` unchanged. Config addition (`error_type_base_url`) is minimal and idiomatic. |
| **II. Testing Standards** | ✅ PASS | All 4 test files identified for update (error_handler_test, error_registry_test, error_dto_test, frontend error tests). Contract test via JSON Schema validation added in quickstart.md. No coverage decrease. |
| **III. UX Consistency** | ✅ PASS | Frontend data-model changes (`web/lib/errors.ts`) are field remapping only — no visual or behavioral changes to modals/toasts/forms. |
| **IV. Performance** | ✅ PASS | Static type URI and title strings stored in `ErrorMapping` struct — zero per-request dynamic allocation for type/title construction. JSON serialization overhead equivalent to current format (same number of fields). |
| **Security** | ✅ PASS | `detail` field sanitization unchanged. `type` URI does not leak internal implementation details (uses configurable base, not code paths). |

**Post-design gate result**: All principles still pass. Design is consistent with constitution.

## Project Structure

### Documentation (this feature)

```text
specs/003-rfc9457-error-format/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0: RFC 9457 structure research
├── data-model.md        # Phase 1: DTO and entity changes
├── quickstart.md        # Phase 1: Validation guide
└── contracts/           # Phase 1: JSON Schema for RFC 9457 response
    └── problem-detail.schema.json
```

### Source Code (affected paths)

```text
# Backend (Go)
internal/
├── api/
│   ├── dto/
│   │   ├── error_dto.go          # REPLACE: new RFC 9457 DTOs
│   │   └── error_dto_test.go     # UPDATE: validate RFC 9457 structure
│   └── middleware/
│       ├── error_handler.go      # UPDATE: serialize RFC 9457 body, Content-Type
│       ├── error_handler_test.go # UPDATE: validate RFC 9457 response
│       ├── error_registry.go     # UPDATE: add type URI slug + title to ErrorMapping
│       └── error_registry_test.go # UPDATE: validate new mapping fields
├── config/
│   └── config.go                 # UPDATE: add ErrorTypeBaseURL field
└── core/
    └── domain/
        └── errors.go             # UNCHANGED (codes remain stable)
configs/
├── config.yaml                   # UPDATE: add error_type_base_url
└── config.yaml.example           # UPDATE: add error_type_base_url example
api/
└── swagger/
    ├── docs.go                   # REGENERATE: updated error response schemas
    ├── swagger.json              # REGENERATE
    └── swagger.yaml              # REGENERATE

# Frontend (TypeScript/Next.js)
web/
├── lib/
│   └── errors.ts                 # UPDATE: parse application/problem+json
├── hooks/
│   └── use-error-modal.tsx       # UPDATE: extract from RFC 9457 fields
├── components/
│   └── layout/
│       └── error-modal.tsx       # UPDATE: display detail/title
└── tests/
    ├── lib/
    │   └── errors.test.ts        # UPDATE: new response format
    └── hooks/
        └── use-error-modal.test.tsx # UPDATE: new response format
```

**Structure Decision**: Web application (Option 2). Backend follows existing Clean Architecture layout; frontend follows Next.js App Router conventions. No new directories or packages — all changes are within existing files and packages.

## Complexity Tracking

No violations to justify. All constitution checks pass.

