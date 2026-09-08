# Quickstart: Runtime Feature Flags

**Date**: 2026-09-08 | **Feature**: `005-feature-flags`

Validation guide — proves the feature works end-to-end. Implementation details live in `tasks.md` and the implementation phase.

## Prerequisites

- Go 1.26+ and the project backend running per `README.md` (`make compose-up` or `make run`)
- PostgreSQL with migrations applied (`make migrate-up`)
- A configured operator username in `configs/config.yaml` under `feature_flags.operators`
- `curl` (or the Postman collection generated from Swagger)

## Config

```yaml
feature_flags:
  environment: "local"          # local | test | staging | production
  refresh_interval: "30s"
  operators:
    - "operator1"
```

---

## Validation Scenarios

### Scenario 1: Create and Toggle a Flag (FR-001, FR-002, FR-005, FR-006)

**Goal**: An operator can create a flag, disable a guarded feature, and enable it without a deployment.

**Steps**:
1. Login as `operator1`; capture the `access_token` cookie.
2. Create the flag:
   ```bash
   curl -X POST http://localhost:8080/admin/feature-flags \
     -H "Authorization: Bearer <operator_token>" \
     -H "Content-Type: application/json" \
     -d '{"key":"beta-editor","name":"Beta Editor","purpose":"Gate the new editor","owner":"team-content","safeDefault":false}'
   ```
   → `201`.
3. Evaluate with a regular user session:
   ```bash
   curl "http://localhost:8080/feature-flags/evaluate?keys=beta-editor" \
     -H "Authorization: Bearer <user_token>"
   ```
   → `enabled:false` (no setting yet → safe default).
4. Enable it:
   ```bash
   curl -X PATCH http://localhost:8080/admin/feature-flags/beta-editor/environment/local \
     -H "Authorization: Bearer <operator_token>" -H "Content-Type: application/json" \
     -d '{"mode":"enabled_all","rolloutPercentage":100,"revision":1,"reason":"Enable for validation"}'
   ```
5. Re-evaluate within 60 s → `enabled:true`.

**Expected Outcome**: `false` before, `true` after, no restart.

### Scenario 2: Gradual Rollout Is Deterministic (FR-007, FR-008, SC-003)

**Goal**: The same user always gets the same result; the audience approximates the percentage.

**Steps**:
1. Set mode to `gradual` with `rolloutPercentage: 50` (revision from previous step).
2. Evaluate `beta-editor` 10 times for the same user → identical result each time.
3. Evaluate across ≥ 20 distinct users → enabled count is within a few points of 50%.

**Expected Outcome**: Repeatable per-user result; aggregate ≈ configured percentage.

### Scenario 3: Explicit Overrides Take Precedence (FR-009, FR-010)

**Goal**: Include forces `true`, exclude forces `false`.

**Steps**:
1. Set `gradual` with `rolloutPercentage: 0` (nobody enabled).
2. Add override: `include: ["alice"]`, `exclude: ["bob"]`.
3. Evaluate as `alice` → `true`; as `bob` → `false`; as `carol` → `false`.

**Expected Outcome**: Overrides win over rollout allocation.

### Scenario 4: Emergency Off Overrides Inclusions (edge case, FR-010)

**Goal**: Global disable beats explicit inclusions.

**Steps**:
1. With `alice` still explicitly included, set mode to `disabled_all`.
2. Evaluate as `alice` → `false`.

**Expected Outcome**: `false` despite inclusion.

### Scenario 5: Conflicting Update Rejected (FR-017)

**Goal**: Two operators editing from the same revision; the later write fails.

**Steps**:
1. Read the setting (revision N).
2. Apply two updates both sending `revision: N` with different reasons.
3. First succeeds (`200`), second returns `409`.

**Expected Outcome**: Second update rejected; operator re-reads (revision N+1) and retries.

### Scenario 6: Archive Is Final (FR-018, FR-019)

**Goal**: Archived flags always safe-default and cannot be edited or re-enabled.

**Steps**:
1. `POST /admin/feature-flags/beta-editor/archive` with a reason.
2. Attempt `PATCH .../environment/local` → `409`.
3. Create a new flag with key `beta-editor` → `409` (key never reused).
4. Evaluate `beta-editor` → `false` (safe default) with `source: "archived"`.

**Expected Outcome**: Archived flag is immutable, key reserved, evaluation pinned to safe default.

### Scenario 7: Failure Falls Back to Safe Default (FR-014, FR-015)

**Goal**: When flag data is unavailable, guarded features degrade safely and unrelated workflows keep working.

**Steps**:
1. Stop the flag data source (e.g., point the evaluator at an unreachable snapshot source, or temporarily drop the cache's data source).
2. Evaluate any flag → `safe_default` result.
3. Confirm the evaluation failure is recorded in logs with flag, environment, and failure category.
4. Confirm unrelated endpoints (e.g., `GET /status`, feed) still respond normally.

**Expected Outcome**: No user workflow breaks because flag data is unreadable.

### Scenario 8: Non-Operator Is Denied (FR-021)

**Goal**: Regular users cannot administer or inspect flag configuration.

**Steps**:
1. As a regular user, call `GET /admin/feature-flags` → `403`.
2. Call `POST /admin/feature-flags` → `403`.
3. Call `GET /feature-flags/evaluate?keys=beta-editor` → `200` with only the boolean decision (no rollout rules, overrides, or audit data).

**Expected Outcome**: Admin surface fully denied; evaluation reveals no configuration.

### Scenario 9: Audit Trail Is Complete (FR-016, SC-007)

**Goal**: Every accepted change is traceable.

**Steps**:
1. Perform a create, a rollout change, an override change, and an archive.
2. `GET /admin/feature-flags/beta-editor/audit` → ordered entries with actor, timestamp, reason, and before/after values for each change.

**Expected Outcome**: Full, immutable history matching every action taken.

---

## Run Commands Summary

```bash
# Backend tests (use case, evaluator, repository integration)
make test
go test ./internal/core/usecase/... ./internal/adapters/featureflag/... ./internal/adapters/repository/... -v

# Frontend tests (feature gate, hook, admin components)
cd web && npm test

# Regenerate docs after adding endpoints
make docs
```
