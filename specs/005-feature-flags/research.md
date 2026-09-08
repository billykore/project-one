# Research: Runtime Feature Flags

**Date**: 2026-09-08 | **Feature**: `005-feature-flags`

## 1. Storage Model

### Decision: Relational tables in the existing PostgreSQL database

Store flags, per-environment settings, user overrides, and audit history in four dedicated tables managed by GORM, alongside the existing migrations.

### Rationale

- **Consistency with the project**: the stack is already PostgreSQL + GORM with granular migrations (`db/migrations/000NNN_*`). Feature flags are configuration state with relationships and audit requirements, not an opaque blob.
- **Queryability**: audit history and per-flag settings need relational queries (list by flag, list by environment, paginate audit). A JSON column would push filtering into application code.
- **Integrity**: database constraints (unique key, unique `(flag_id, environment)`, unique override tuple, FK cascade) enforce FR-018 and FR-020 at the source of truth.
- **Operational simplicity**: no new infrastructure. Operators already back up and operate PostgreSQL.

### Alternatives Considered

| Alternative | Why Rejected |
|-------------|--------------|
| JSON/`jsonb` column in a single `feature_flags` row | Loses referential integrity and per-row audit; complicates indexing and pagination. |
| External flag service (LaunchDarkly, Unleash, OpenFeature provider) | New runtime dependency and network hop; conflicts with the "no new dependencies" goal and the 60-second-local propagation requirement; overkill for boolean flags in the first release. |
| Config file flags only | Requires redeploy to change (violates FR-005); no audit history or operator UI. |

---

## 2. Deterministic Rollout Algorithm

### Decision: FNV-1a hash of `flagKey + environment + username` mapped to a 0–99 bucket

```go
h := fnv.New64a()
h.Write([]byte(flagKey + "|" + environment + "|" + username))
bucket := int(h.Sum64() % 100) // 0..99
enabled := bucket < rolloutPercentage
```

### Rationale

- **Deterministic and stable**: the same `(flag, environment, user)` always lands in the same bucket across processes and restarts, satisfying FR-008 and SC-003 (100% repeatability).
- **Standard library only**: `hash/fnv` ships with Go; no third-party dependency, consistent with the project's minimal-dependency philosophy.
- **Uniform distribution**: FNV-1a distributes identifiers evenly enough that bucket assignment is effectively uniform, satisfying SC-004 (enabled population within 5 percentage points for audiences ≥ 10,000).

### Alternatives Considered

| Alternative | Why Rejected |
|-------------|--------------|
| Hash of username alone | Same user would always get the same bucket across all flags, causing correlated rollouts (unwanted). |
| Random with stored assignment | Requires persisting a per-user assignment on first evaluation; extra table and consistency burden. |
| Murmur3 / xxhash | Equivalent quality but would require a dependency; FNV-1a is sufficient for bucketing. |

**Note on percentage change stability (spec edge case)**: because the bucket is derived from the key+environment+user and compared against a threshold, raising the percentage keeps previously-enabled users enabled; lowering it removes users above the new threshold. This matches the documented edge-case behavior.

---

## 3. Caching & Propagation (≤ 60 s)

### Decision: In-memory snapshot cache with periodic refresh and write-triggered invalidation

- The evaluator adapter (`internal/adapters/featureflag/evaluator.go`) holds a `map[key]environmentSetting` snapshot guarded by `sync.RWMutex`.
- A background goroutine refreshes the snapshot every 30 seconds by loading all flags/settings/overrides for the configured environment (a small bounded query set).
- Every successful admin write calls `Refresh` immediately, so changes propagate in well under 60 seconds in the common case.
- On snapshot load failure at startup, the evaluator serves safe defaults and retries on the next tick, logging the failure (FR-014/FR-015).

### Rationale

- **Hot path performance**: evaluation is an O(1) in-memory lookup with no database call, meeting the <1 ms goal and constitution read-latency limits.
- **60-second guarantee**: 30-second polling plus immediate invalidation comfortably satisfies FR-013.
- **Graceful degradation**: a cache that falls back to safe defaults keeps unrelated features running when flag data is unavailable (FR-014).

### Alternatives Considered

| Alternative | Why Rejected |
|-------------|--------------|
| Evaluate directly against DB per request | Adds a query to every guarded request; risks blocking the request path. |
| Push-based invalidation only (no polling) | A missed invalidation message would leave a stale decision indefinitely; polling provides a safety net. |
| External config center with webhooks | New infrastructure and network dependency; overkill for this scale. |

---

## 4. Operator Authorization

### Decision: Config-driven operator allowlist

Add `feature_flags.operators` (list of usernames) to `configs/config.yaml`. An `OperatorOnly` middleware runs after `Authorize` and rejects requests whose authenticated username is not in the allowlist (403, no flag detail disclosed).

### Rationale

- **No schema change**: the project has no role model; introducing one for the first release would touch `users`, tokens, and every protected handler.
- **Minimal and auditable**: the allowlist is versioned in config, small, and easy to review.
- **Meets FR-001/FR-021**: only listed operators can administer flags; ordinary users are denied without revealing configuration.

### Alternatives Considered

| Alternative | Why Rejected |
|-------------|--------------|
| `is_operator` column on users + migration | Requires a role model, admin assignment flow, and broader security review for a first release. |
| Reuse of an existing "admin" concept | None exists in the codebase. |
| Public-key or token-based admin auth | Adds a parallel auth mechanism and key-management burden. |

**Assumption documented in the spec**: a full role/permission model may replace the allowlist in a later feature.

---

## 5. Case-Insensitive Unique Flag Keys

### Decision: Unique index on `lower(trim(key))`

```sql
CREATE UNIQUE INDEX ux_feature_flags_key ON feature_flags (lower(trim(key)));
```

### Rationale

- Enforces FR-018 (unique after trimming whitespace and ignoring case) at the database level, immune to races.
- Application code still normalizes keys before insert for a clean error message (422 on duplicate).

### Alternatives Considered

| Alternative | Why Rejected |
|-------------|--------------|
| `citext` column type | Requires the `citext` extension; a functional index is dependency-free and equally effective. |
| Application-level uniqueness check only | Race condition between check and insert. |

---

## 6. Optimistic Concurrency

### Decision: Revision column on `feature_flag_settings`

Each environment setting carries a monotonically increasing `revision`. Updates must supply the `revision` they started from; a mismatch returns a 409 conflict (FR-017) telling the operator to re-read the current state.

### Rationale

- Detects the "two operators edit the same flag" edge case without long-lived locks or transactions.
- Simple to implement and matches the spec's "rejected and retried against current state" requirement.

### Alternatives Considered

| Alternative | Why Rejected |
|-------------|--------------|
| Pessimistic row lock during edit | Blocks operators and requires a session-scoped lock lifecycle. |
| Timestamp comparison | Clock skew between app instances can mask conflicts. |

---

## 7. Evaluation Precedence

### Decision: First-match-wins order (mirrors FR-010)

1. **Archived** → declared safe default (FR-019)
2. **Mode = disabled-for-everyone** → `false` (emergency off; overrides inclusions, per edge case)
3. **Explicit exclusion** for the user → `false`
4. **Explicit inclusion** for the user → `true`
5. **Mode = gradual rollout** → FNV-1a bucket `< percentage`
6. **Mode = enabled-for-everyone** → `true`
7. **No setting / evaluation error** → declared safe default

Anonymous visitors (no username) skip steps 3–4 and, in gradual mode, receive the safe default (FR-011); they are never assigned a new identity.

### Rationale

- Encodes the spec's precedence and both edge cases (global off overrides inclusions; archived always safe-defaults) in a single, unit-testable function.
- Each step is a trivial boolean check, keeping evaluation O(1).

---

## 8. Frontend Integration

### Decision: Server-side evaluation via BFF, plus a thin client gate

- The Go API exposes `GET /feature-flags/evaluate?keys=a,b,c` (optionally authenticated), returning per-key boolean decisions computed by the evaluator.
- Next.js route handlers proxy evaluation (`web/app/api/feature-flags/route.ts`) and admin CRUD (`web/app/api/admin/feature-flags/*`), forwarding the session cookie to the Go API via the existing server-only `API_URL` pattern.
- A `use-feature-flag` hook + `FeatureGate` component read evaluated flags from the BFF to hide guarded UI; server components can evaluate directly during render.
- The operator dashboard (`web/app/admin/feature-flags/page.tsx`) renders the admin API through the BFF.

### Rationale

- Backend gating of direct actions (FR-012) is authoritative; the frontend gate only controls presentation.
- Reuses the existing BFF cookie-forwarding architecture (`web/lib/server-fetch.ts`).

### Alternatives Considered

| Alternative | Why Rejected |
|-------------|--------------|
| Expose raw flag config to the browser | Leaks rollout rules and audit data to ordinary users (violates FR-021). |
| Evaluate flags in the browser from a public endpoint only | Can't be trusted for protected actions; still needs backend enforcement. |

---

## 9. Anti-Patterns Avoided

- **No flag evaluation on the database hot path** — evaluation reads from memory only.
- **No unbounded queries** — snapshot load is bounded by flag count; audit list is paginated.
- **No leaking of flag internals to non-operators** — evaluate endpoint returns only booleans for the requested keys; admin endpoints are allowlist-protected.
- **No new third-party dependencies** — deterministic hashing and caching use the standard library.
- **No scheduled activation, multivariate values, or experiment analysis** — explicitly out of scope (FR-003, FR-022).
