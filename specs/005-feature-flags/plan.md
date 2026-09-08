# Implementation Plan: Runtime Feature Flags

**Branch**: `feat/feature-flags` | **Date**: 2026-09-08 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/005-feature-flags/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Add a runtime feature flag capability so authorized operators can toggle named boolean features per environment, run percentage rollouts with per-user overrides, and audit every change — without redeploying. The Go API gains a feature-flag domain (entity, ports, use case), a PostgreSQL persistence layer, an in-memory cache evaluator with deterministic FNV-1a rollout assignment, and admin plus evaluation endpoints. The Next.js frontend gains a minimal operator dashboard, BFF proxies, and a feature gate for guarded UI.

## Technical Context

**Language/Version**: Go 1.26.2 (backend); TypeScript 5 / Next.js 16.2.9 / React 19.2.4 (frontend)

**Primary Dependencies**: Existing stack only — Echo 4.15, GORM, Viper, `validator/v10`, `log/slog`. No new third-party dependencies; deterministic rollout uses the Go standard library `hash/fnv`.

**Storage**: PostgreSQL — 4 new tables: `feature_flags`, `feature_flag_settings`, `feature_flag_overrides`, `feature_flag_audit_logs`.

**Testing**: `go test` with GoMock-generated mocks (`make mock`), repository integration tests against the test database, Vitest for the frontend gate/hook and admin components.

**Target Platform**: Linux server (Docker deployment via `deployments/compose.yml`).

**Project Type**: Web service backend (Go API) + Next.js frontend (BFF route handlers, App Router).

**Performance Goals**: In-memory flag evaluation < 1 ms per request; admin read endpoints p95 < 200 ms and write endpoints p95 < 500 ms (per constitution); flag changes effective within 60 s.

**Constraints**: Strict Clean Architecture layer separation; evaluation must never block the request path and must fall back to the declared safe default on failure; RFC 9457 error responses; Swagger annotations on new endpoints; no deployment/restart required for flag changes.

**Scale/Scope**: 4 new tables; 1 new domain entity cluster (flag, setting, override, audit record, decision); 1 repository, 1 evaluator cache, 1 use case, 1 handler, 2 middlewares; 8 admin endpoints + 1 evaluation endpoint; a minimal operator dashboard and a feature gate in the frontend.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Code Quality & Clean Architecture** | ✅ PASS | Domain entities (`internal/core/domain/feature_flag.go`) are pure Go with no framework imports; ports (`internal/core/ports/feature_flag.go`) define repository/evaluator/use-case interfaces; use case orchestrates against ports only; GORM repository lives in `internal/adapters/repository/`; Echo handler + middleware in `internal/api/`. |
| **II. Testing Standards** | ✅ PASS | Use-case unit tests with GoMock; repository integration tests; evaluator unit tests for precedence/determinism/failure fallback; Vitest for the frontend gate and hook. `make mock` regenerated after interface changes. |
| **III. User Experience Consistency** | ✅ PASS | Admin dashboard reuses the glassmorphism design system and shadcn/ui primitives in `web/components/`; loading/empty/error/success states included; regular user-facing pages are unchanged except where a guarded feature is explicitly gated. |
| **IV. Performance Requirements** | ✅ PASS | Evaluation is an in-memory O(1) map lookup + O(1) hash, no DB call on the hot path; snapshot load is a small bounded query set; audit list endpoint paginated; indexes on `key`, `(flag_id, environment)`, and audit `(flag_id, created_at)`. |
| **Security & Authentication** | ✅ PASS | Admin endpoints require `Authorize` + operator allowlist middleware; audit records are insert-only; all inputs validated server-side with `validator/v10`; ordinary users cannot read or mutate flag data. |
| **Development Workflow** | ✅ PASS | Conventional commits; `make check` must pass; branch `feat/feature-flags` follows the `###-short-description` convention; new endpoints documented with Swagger annotations (`make docs`). |

**Gate Result**: ✅ All principles satisfied. No violations to justify.

## Project Structure

### Documentation (this feature)

```text
specs/005-feature-flags/
├── plan.md              # This file
├── research.md          # Phase 0: storage, rollout, caching, authorization decisions
├── data-model.md        # Phase 1: entities, tables, migrations, state transitions
├── quickstart.md        # Phase 1: setup and validation scenarios
├── contracts/           # Phase 1: admin + evaluation API contract
│   └── feature-flags-api.md
└── tasks.md             # Phase 2: (/speckit.tasks command - NOT created here)
```

### Source Code (repository root)

```text
internal/
├── core/
│   ├── domain/
│   │   ├── feature_flag.go          # NEW: FeatureFlag, EnvironmentSetting, UserOverride,
│   │   │                             #       AuditRecord, FeatureFlagDecision + enums
│   │   └── errors.go                # MODIFIED: add feature-flag sentinel errors
│   ├── ports/
│   │   ├── feature_flag.go          # NEW: FeatureFlagRepository, FeatureFlagEvaluator,
│   │   │                             #       FeatureFlagUseCase interfaces
│   │   └── mocks/                   # REGENERATED: make mock
│   └── usecase/
│       ├── feature_flag_usecase.go  # NEW: admin CRUD, validation, audit, optimistic concurrency
│       └── feature_flag_usecase_test.go  # NEW
├── adapters/
│   ├── repository/
│   │   └── feature_flag_repository.go    # NEW: GORM persistence + snapshot load
│   ├── featureflag/
│   │   ├── evaluator.go             # NEW: in-memory cache + FNV-1a rollout evaluator
│   │   └── evaluator_test.go        # NEW
├── config/
│   └── config.go                    # MODIFIED: add FeatureFlagsConfig
└── api/
    ├── handler/
    │   ├── feature_flag_handler.go  # NEW: admin CRUD + evaluate endpoints
    │   └── feature_flag_handler_test.go  # NEW
    └── middleware/
        ├── feature_flag_operator.go # NEW: operator allowlist middleware
        └── feature_flag_gate.go     # NEW: per-route guarded-feature middleware

cmd/
└── main.go                          # MODIFIED: wire repository, use case, evaluator, handler, middlewares

configs/
├── config.yaml                      # MODIFIED: feature_flags section
└── config.yaml.example              # MODIFIED

db/migrations/
├── 000018_create_feature_flags_table.up.sql / .down.sql
├── 000019_create_feature_flag_settings_table.up.sql / .down.sql
├── 000020_create_feature_flag_overrides_table.up.sql / .down.sql
└── 000021_create_feature_flag_audit_logs_table.up.sql / .down.sql

web/
├── app/
│   ├── admin/feature-flags/page.tsx # NEW: operator dashboard (list, detail, toggle, rollout, archive)
│   └── api/
│       ├── feature-flags/route.ts   # NEW: BFF evaluate proxy
│       └── admin/feature-flags/     # NEW: BFF admin proxies (CRUD)
├── components/
│   └── feature-flag/
│       └── feature-gate.tsx         # NEW: client feature gate component
├── hooks/
│   └── use-feature-flag.ts          # NEW: client evaluation hook
└── lib/
    └── feature-flags-api.ts         # NEW: BFF client helpers
```

**Structure Decision**: Single repo, two-project layout (Go API + Next.js frontend). Feature-flag domain, ports, use case, and adapters follow the existing Clean Architecture packages; the evaluator cache is colocated in `internal/adapters/featureflag/` alongside other adapter implementations. Four granular migrations (one per table) match the existing `db/migrations/` convention. Frontend pieces live under `web/app`, `web/components`, `web/hooks`, and `web/lib` following the established Next.js App Router structure.

## Complexity Tracking

> No constitution violations — this section is empty.

## Phase Outputs

### Phase 0: Research — [research.md](./research.md)

- Relational storage (4 tables) over JSON blob or external flag service
- Deterministic rollout via FNV-1a bucket assignment (standard library, stable)
- In-memory snapshot cache with 30 s refresh + write-triggered invalidation (≤60 s propagation)
- Config-driven operator allowlist over a DB role model
- Case-insensitive unique keys via `lower(trim(key))` unique index
- Revision-based optimistic concurrency on environment settings
- Precedence order matching FR-010 with safe-default fallback

### Phase 1: Design — [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

- `domain.FeatureFlag`, `EnvironmentSetting`, `UserOverride`, `AuditRecord`, `FeatureFlagDecision`
- 4 migrations with indexes and constraints
- Admin + evaluation API contract with RFC 9457 errors
- Validation scenarios: toggle, rollout determinism, overrides, emergency off, archival, audit, failure fallback

### Phase 2: Tasks — to be generated by `/speckit.tasks`
