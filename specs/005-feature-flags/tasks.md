# Tasks: Runtime Feature Flags

**Input**: Design documents from `/specs/005-feature-flags/`

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅, quickstart.md ✅

**Tests**: Included — the project constitution mandates unit tests for all use cases and adapters, and Vitest coverage for frontend utilities. Test tasks are written first and must fail before implementation.

**Organization**: Tasks are grouped by user story so each story can be implemented, tested, and delivered independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Exact file paths included in every task

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Configuration plumbing for the feature-flag capability.

- [ ] T001 Add `FeatureFlagsConfig` struct (fields: `environment`, `refresh_interval`, `operators`) to `internal/config/config.go`; add defaults, `BindEnv` entries, and validation in `Load`; extend `internal/config/config_test.go`
- [ ] T002 Add `feature_flags` section (`environment`, `refresh_interval`, `operators`) to `configs/config.yaml` and `configs/config.yaml.example`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Domain model, ports, persistence, and the evaluation engine — shared by every user story.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T003 [P] Create domain entities and enums (`FeatureFlag`, `EnvironmentSetting`, `UserOverride`, `AuditRecord`, `FeatureFlagDecision`; `LifecycleState`, `Environment`, `AvailabilityMode`, `OverrideType`) in `internal/core/domain/feature_flag.go`
- [ ] T004 [P] Add feature-flag sentinel errors (e.g., `ErrFlagNotFound`, `ErrFlagKeyExists`, `ErrFlagArchived`, `ErrRevisionConflict`, `ErrInvalidMode`) to `internal/core/domain/errors.go`
- [ ] T005 [P] Create migrations `000018_create_feature_flags_table`, `000019_create_feature_flag_settings_table`, `000020_create_feature_flag_overrides_table`, `000021_create_feature_flag_audit_logs_table` (up + down files) in `db/migrations/`
- [ ] T006 Define `FeatureFlagRepository`, `FeatureFlagEvaluator`, and `FeatureFlagUseCase` interfaces in `internal/core/ports/feature_flag.go`
- [ ] T007 Implement GORM `FeatureFlagRepository` in `internal/adapters/repository/feature_flag_repository.go` (flag/setting/override CRUD, `LoadSnapshot` for the evaluator, revision-based optimistic concurrency, audit append)
- [ ] T008 Implement in-memory evaluator in `internal/adapters/featureflag/evaluator.go` (FNV-1a bucket assignment, FR-010 precedence, 30s refresh + write-triggered invalidation, safe-default fallback on failure, and structured `slog` recording of evaluation failures with flag, environment, time, and failure category — FR-015)
- [ ] T009 Run `make mock` to regenerate `internal/core/ports/mocks/`

**Checkpoint**: Foundation ready — user story implementation can now begin.

---

## Phase 3: User Story 1 - Safely Control Feature Availability (Priority: P1) 🎯 MVP

**Goal**: Authorized operators can turn a named boolean feature on or off per environment without a deployment; evaluation falls back to the declared safe default on failure.

**Independent Test**: Create a flag, confirm the feature is unavailable while disabled, enable it and confirm availability within 60 s, then disable it and confirm access is withdrawn — no restart.

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation.**

- [ ] T010 [P] [US1] Unit tests for flag create/toggle/evaluate, validation, and safe-default behavior in `internal/core/usecase/feature_flag_usecase_test.go`
- [ ] T011 [P] [US1] Unit tests for evaluator precedence (enabled/disabled/safe default), failure fallback, and cache refresh/invalidation propagation within the refresh interval (SC-002) in `internal/adapters/featureflag/evaluator_test.go`
- [ ] T012 [P] [US1] Integration tests for repository persistence and optimistic concurrency in `internal/adapters/repository/feature_flag_repository_test.go`

### Implementation for User Story 1

- [ ] T013 [US1] Implement `FeatureFlagUseCase` (create, get, list, set availability `enabled_all`/`disabled_all`, evaluate with safe default, validation) in `internal/core/usecase/feature_flag_usecase.go`
- [ ] T014 [US1] Implement `OperatorOnly` middleware (config allowlist check after `Authorize`) in `internal/api/middleware/feature_flag_operator.go`
- [ ] T015 [US1] Implement `FeatureFlagHandler` (create, list, get, patch environment) with Swagger annotations in `internal/api/handler/feature_flag_handler.go`
- [ ] T016 [US1] Implement `GET /feature-flags/evaluate?keys=` endpoint (per-user or anonymous) in `internal/api/handler/feature_flag_handler.go`
- [ ] T017 [US1] Implement `FeatureFlagGate` route middleware (guarded-feature check with safe default) in `internal/api/middleware/feature_flag_gate.go`, and apply it to one existing route with a matching use-case-level guard check to demonstrate both entry-point and direct-action protection (FR-012)
- [ ] T018 [US1] Wire repository, evaluator, use case, handler, and middlewares into `newApplication` and `registerRoutes` in `cmd/main.go`
- [ ] T019 [US1] Add BFF evaluate proxy `web/app/api/feature-flags/route.ts` and client helpers in `web/lib/feature-flags-api.ts`
- [ ] T020 [US1] Add `FeatureGate` component in `web/components/feature-flag/feature-gate.tsx` and `use-feature-flag` hook in `web/hooks/use-feature-flag.ts`
- [ ] T021 [US1] Add minimal operator dashboard (list flags + toggle enabled/disabled) in `web/app/admin/feature-flags/page.tsx` with BFF admin proxies under `web/app/api/admin/feature-flags/`

**Checkpoint**: User Story 1 fully functional and independently testable — operators can enable/disable features per environment and guard routes/UI.

---

## Phase 4: User Story 2 - Release Gradually to a Stable Audience (Priority: P2)

**Goal**: Operators can roll a feature out to a percentage of signed-in users and explicitly include or exclude selected users.

**Independent Test**: Configure a partial rollout, evaluate repeatedly for the same users and verify stable results, then apply include/exclude overrides and verify they take precedence.

### Tests for User Story 2 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation.**

- [ ] T022 [P] [US2] Unit tests for gradual-rollout allocation and override precedence in `internal/core/usecase/feature_flag_usecase_test.go`
- [ ] T023 [P] [US2] Unit tests for rollout determinism (same user, stable result), override precedence, and bucket-distribution accuracy within 5 percentage points over a simulated 10,000-user audience (SC-004) in `internal/adapters/featureflag/evaluator_test.go`

### Implementation for User Story 2

- [ ] T024 [US2] Extend `FeatureFlagUseCase` with `gradual` mode and include/exclude override management (conflicting override validation) in `internal/core/usecase/feature_flag_usecase.go`
- [ ] T025 [US2] Add `PUT /admin/feature-flags/{key}/overrides` endpoint with Swagger annotation in `internal/api/handler/feature_flag_handler.go`
- [ ] T026 [US2] Add rollout-percentage and override controls to the operator dashboard in `web/app/admin/feature-flags/page.tsx`

**Checkpoint**: User Stories 1 AND 2 work independently — partial rollouts and overrides are functional on top of the on/off control.

---

## Phase 5: User Story 3 - Govern and Audit Flag Changes (Priority: P3)

**Goal**: Operators can see each flag's metadata, lifecycle state, and full change history, and archive flags permanently.

**Independent Test**: Create a flag, change its settings, archive it, and verify the full ordered history identifies each change, actor, time, and reason.

### Tests for User Story 3 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation.**

- [ ] T027 [P] [US3] Unit tests for archive transitions and audit-record creation in `internal/core/usecase/feature_flag_usecase_test.go`

### Implementation for User Story 3

- [ ] T028 [US3] Extend `FeatureFlagUseCase` with archive transitions and immutable audit-record writing (previous/new value, actor, reason) in `internal/core/usecase/feature_flag_usecase.go`
- [ ] T029 [US3] Add `POST /admin/feature-flags/{key}/archive` and `GET /admin/feature-flags/{key}/audit` endpoints with Swagger annotations in `internal/api/handler/feature_flag_handler.go`
- [ ] T030 [US3] Add archive action and audit-history view to the operator dashboard in `web/app/admin/feature-flags/page.tsx`

**Checkpoint**: All user stories independently functional — flags are governable and auditable end-to-end.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Hardening and validation across all stories.

- [ ] T031 [P] Add handler unit tests (request validation, authorization, error mapping) in `internal/api/handler/feature_flag_handler_test.go`
- [ ] T032 [P] Add Vitest tests for `FeatureGate` and `use-feature-flag` in `web/tests/`
- [ ] T033 Run `make check` (vet + lint + test), `make docs` to regenerate Swagger, and `npm test` in `web/`; fix all failures
- [ ] T034 Execute the validation scenarios in `specs/005-feature-flags/quickstart.md` and confirm each expected outcome

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately.
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories.
- **User Stories (Phases 3–5)**: All depend on Phase 2 completion. US2 and US3 extend the use case and handler created in US1, so they are sequential in priority order (P1 → P2 → P3) when implemented by one engineer.
- **Polish (Phase 6)**: Depends on all desired user stories.

### User Story Dependencies

- **User Story 1 (P1)**: Starts after Phase 2. No dependencies on other stories.
- **User Story 2 (P2)**: Starts after US1 — extends `feature_flag_usecase.go` and the admin dashboard.
- **User Story 3 (P3)**: Starts after US2 — extends the same use case and handler files.

### Within Each User Story

- Tests MUST be written and FAIL before implementation.
- Domain/ports before repository/evaluator before use case before handler.
- Backend endpoints before frontend UI.
- Story complete before moving to the next priority.

### Parallel Opportunities

- Phase 2: T003, T004, T005 are independent files and can run in parallel.
- US1 tests: T010, T011, T012 are independent and can run in parallel.
- US2 tests: T022, T023 are independent and can run in parallel.
- Polish: T031, T032 are independent and can run in parallel.

---

## Parallel Example: Phase 2 Foundation

```bash
# Launch independent foundational tasks together:
Task: "Create domain entities and enums in internal/core/domain/feature_flag.go"
Task: "Add feature-flag sentinel errors to internal/core/domain/errors.go"
Task: "Create migrations 000018–000021 in db/migrations/"
```

## Parallel Example: User Story 1

```bash
# Launch all US1 tests together (they fail before implementation):
Task: "Unit tests for flag create/toggle/evaluate in internal/core/usecase/feature_flag_usecase_test.go"
Task: "Unit tests for evaluator precedence in internal/adapters/featureflag/evaluator_test.go"
Task: "Integration tests for repository persistence in internal/adapters/repository/feature_flag_repository_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: test US1 independently (create/toggle/evaluate per quickstart Scenarios 1 and 7).
5. Deploy/demo if ready.

### Incremental Delivery

1. Setup + Foundational → foundation ready.
2. Add US1 → test independently → deploy/demo (MVP: on/off control + safe default).
3. Add US2 → test independently → deploy/demo (gradual rollout + overrides).
4. Add US3 → test independently → deploy/demo (audit + archival).
5. Polish: full `make check`, docs, quickstart validation.

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together.
2. Once Foundational is done:
   - Developer A: User Story 1 (use case + handler + wiring).
   - Developer B: User Story 2 (rollout + overrides) after US1 use case is stubbed.
   - Developer C: User Story 3 (audit + archive) after US1 use case is stubbed.
3. Since US2/US3 extend the same files, coordinate merges or split by file where possible.

---

## Notes

- [P] tasks = different files, no dependencies.
- [Story] label maps each task to its user story for traceability.
- Verify tests fail before implementing; commit after each task or logical group.
- Backend gating (gate middleware + use-case checks) is authoritative; the frontend `FeatureGate` only controls presentation.
