# Design Quality Checklist: Global Error Handling Middleware

**Purpose**: Validate requirements quality across API contract, middleware behavior, and handler refactoring safety
**Created**: 2026-07-13
**Depth**: Standard PR review gate
**Focus**: API Contract + Middleware Behavior + Refactoring Safety (balanced)
**Mandatory Gate**: No production data leakage
**Feature**: [spec.md](../spec.md) | [plan.md](../plan.md) | [data-model.md](../data-model.md) | [research.md](../research.md)

---

## API Contract Quality

- [ ] CHK001 — Are all 17 domain sentinel errors assigned a unique, documented machine-readable error code? [Completeness, Data Model §Mapping Table]
- [ ] CHK002 — Are HTTP status code mappings explicitly defined for every sentinel error in a single authoritative source? [Completeness, Spec §FR-002, Data Model §Mapping Table]
- [ ] CHK003 — Is the error response JSON structure fully specified with field types, required/optional declarations, and value constraints for every field? [Completeness, Spec §FR-004, contracts/error-response.schema.json]
- [ ] CHK004 — Are error codes documented as stable identifiers (not subject to renaming without a breaking-change notice)? [Clarity, Spec §US2-A3 — frontend consumers build dispatch logic on codes]
- [ ] CHK005 — Is the `details` array schema defined for all cardinalities: empty (non-validation errors), single-field, and multi-field validation failures? [Completeness, Data Model §ErrorDetail]
- [ ] CHK006 — Is the UPPER_SNAKE_CASE naming convention for error codes explicitly documented so future contributors follow the same pattern? [Clarity, Research §4, Data Model §ErrorCode]
- [ ] CHK007 — Are error code uniqueness constraints defined — can two sentinel errors share the same code, or must each code be unique? [Completeness, Data Model §ErrorCode]
- [ ] CHK008 — Is the `request_id` field's source (Echo RequestID middleware) and its format explicitly specified? [Clarity, Spec §FR-004, Spec §US1-A6]
- [ ] CHK009 — Is the `details[].field` naming convention specified — should it use JSON tag names (snake_case) or Go struct field names (PascalCase)? [Clarity, Data Model §ErrorDetail — says "JSON tag name" but needs explicit rule]
- [ ] CHK010 — Is backward compatibility with the current `dto.ErrorResponse` format addressed — must existing API consumers migrate, or is there a deprecation window? [Gap, Spec §Assumptions — frontend update out of scope but no migration path defined]

---

## Middleware Behavior

- [ ] CHK011 — Is the error chain resolution order specified when an error wraps multiple sentinels (e.g., `fmt.Errorf("a: %w", fmt.Errorf("b: %w", ErrPostNotFound))` where both intermediate and leaf could match)? [Clarity, Research §2 — says "first recognized" but what determines iteration order of map?]
- [ ] CHK012 — Is the two-layer architecture contract (middleware wrapper → `c.Error(err)` → custom `HTTPErrorHandler`) defined with clear responsibility boundaries for each layer? [Clarity, Plan §Summary, Research §1]
- [ ] CHK013 — Are logging level assignments explicitly specified per HTTP status class (4xx → Warn, 5xx → Error)? [Completeness, Spec §FR-005]
- [ ] CHK014 — Are the required structured log fields (`request_id`, `method`, `path`, `status`, `error_code`, `error`, `user`) each defined with their source and type? [Clarity, Spec §US3-A1]
- [ ] CHK015 — Is the `user` log field behavior defined when the request has no authentication context (anonymous/guest requests)? [Completeness, Spec §US3-A1 — says "anonymous" but needs explicit mapping rule]
- [ ] CHK016 — Is the stack trace capture mechanism's behavior explicitly differentiated between production (omitted) and non-production (included via `runtime/debug.Stack()`)? [Clarity, Spec §US3-A2, US3-A3, Research §5]
- [ ] CHK017 — Is the `Content-Type: application/json` header requirement defined for ALL error responses, including those that may originate from non-JSON endpoints? [Completeness, Spec §FR-006 — what about WebSocket upgrade errors?]
- [ ] CHK018 — Are requirements defined for what happens when the error middleware itself panics or errors during error processing? [Gap, Edge Case — recursive failure scenario]
- [ ] CHK019 — Is the success-path behavior (nil error → middleware is a no-op) explicitly specified and testable? [Clarity, Spec §FR-007, Spec §US1-A5]

---

## Production Data Leakage Prevention ⚠️ MANDATORY GATE

- [ ] CHK020 — Is "production mode" detection mechanism explicitly specified — which config field, what string value, and is the comparison case-sensitive? [Clarity, Spec §Assumptions — says `config.App.Env == "production"` but needs exact source field path]
- [ ] CHK021 — Are ALL categories of information that MUST be sanitized in production exhaustively enumerated: stack traces, raw `err.Error()` strings, SQL error messages, file paths, panic messages, internal type names? [Completeness, Spec §FR-010, Spec §Edge Cases]
- [ ] CHK022 — Is the sanitization behavior specified when a sentinel error's registered default message itself inadvertently contains internal details (e.g., a message added by a developer that includes a table name)? [Edge Case, Spec §Edge Cases — says "developers are responsible" but no automated guard]
- [ ] CHK023 — Is the allowlist for production-safe response fields explicitly defined: error code (always), sanitized message (always), request_id (always), field details (only for validation errors)? [Completeness, Spec §FR-010]
- [ ] CHK024 — Are requirements defined for logging sensitive data — does the server-side log entry ever contain PII or secrets that should not appear in log aggregation systems? [Completeness, Spec §US3-A1 — log fields include "user" and "error" but PII/secret filtering is not addressed]
- [ ] CHK025 — Is there a requirement that the sanitization behavior is testable — can a test verify that in production mode, the response body contains no substring from the raw Go error? [Measurability, Spec §SC-005]

---

## Handler Refactoring Safety

- [ ] CHK026 — Is the phased migration strategy documented with explicit ordering: middleware deployed first (dormant), then handlers refactored one file at a time? [Clarity, Plan §Summary, Research §6]
- [ ] CHK027 — Are explicit requirements defined for preserving semantically equivalent HTTP status codes across all refactored endpoints? [Completeness, Spec §FR-012, Spec §SC-006]
- [ ] CHK028 — Is the success-path behavior explicitly preserved in requirements — refactored handlers still call `c.JSON(status, data)` for 2xx responses? [Clarity, Research §6 — stated in rationale but not as an FR]
- [ ] CHK029 — Are rollback/revert requirements defined if the refactored middleware causes unexpected behavior in production? [Gap, Coverage — no rollback strategy documented]
- [ ] CHK030 — Is there a requirement that the handler refactoring must not change any existing test assertions about HTTP status codes? [Consistency, Spec §SC-006]
- [ ] CHK031 — Are the `post_handler.go` logging gaps (no logger field, no error logging) explicitly addressed in the refactoring requirements? [Completeness, Plan §Project Structure — notes adding logger but no FR]
- [ ] CHK032 — Is the custom HTTP status override mechanism (FR-008) specified clearly enough that handlers can use it without reintroducing boilerplate? [Clarity, Spec §FR-008 — says "expose a way" but mechanism is only in Research §6]

---

## Cross-Cutting Consistency

- [ ] CHK033 — Are the error response requirements consistent between spec §FR-004, the JSON schema (contracts/error-response.schema.json), and data-model.md §ErrorResponse? [Consistency]
- [ ] CHK034 — Do the HTTP status code mappings in data-model.md align with the status codes currently returned by handlers for each sentinel error? [Consistency, Spec §SC-006]
- [ ] CHK035 — Is the middleware registration order requirement (after RequestID and Recover) consistent with the dependency on RequestID for the `request_id` field? [Consistency, Spec §FR-011]
- [ ] CHK036 — Are error code constants defined in the same location as sentinel errors (`internal/core/domain/errors.go`) consistent with Clean Architecture principles (no framework imports in domain)? [Consistency, Plan §Constitution Check §I]

---

## Dependencies & Assumptions Validation

- [ ] CHK037 — Is the assumption that Echo's RequestID middleware always runs before the error middleware validated or enforced? [Assumption, Spec §Assumptions]
- [ ] CHK038 — Is the dependency on `ports.Logger` interface documented with its full method set (Debug, Info, Warn, Error, Fatal), and is the adapter using `log/slog` compatible? [Dependency, Plan §Technical Context]
- [ ] CHK039 — Is the assumption that existing sentinel errors will not be renamed during this feature explicitly documented? [Assumption, Spec §Assumptions]
- [ ] CHK040 — Are the frontend adaptation requirements explicitly scoped as out of scope for this feature with a clear boundary? [Scope, Spec §Assumptions]

---

## Notes

- **Mandatory Gate**: Items CHK020–CHK025 (Production Data Leakage Prevention) are blocking — all must pass before merge.
- **Priority**: CHK001–CHK010 (API Contract) form the stable public interface — resolve these first as they have the widest blast radius.
- **Gap items**: CHK010 (backward compatibility), CHK018 (recursive failure), CHK029 (rollback strategy) identify missing requirements that may need spec amendments.
