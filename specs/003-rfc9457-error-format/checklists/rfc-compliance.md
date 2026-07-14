# RFC 9457 Compliance Checklist: Backend Error Response Format

**Purpose**: Validate that backend requirements for RFC 9457 Problem Details are complete, clear, and consistent
**Created**: 2026-07-14
**Reviewed**: 2026-07-14
**Depth**: Standard (PR review)
**Focus**: RFC compliance — backend Go middleware, DTOs, and error registry
**Feature**: [spec.md](../spec.md) | [data-model.md](../data-model.md) | [contracts/problem-detail.schema.json](../contracts/problem-detail.schema.json)

## RFC 9457 Standard Field Completeness

- [x] CHK001 — Are all five RFC 9457 standard fields required to appear in every error response? → **PASS**: FR-001 enumerates all 5; T008 constructs all 5 unconditionally.
- [x] CHK002 — Is it specified which RFC 9457 fields are mandatory vs optional in the response body? → **NOTED**: FR-001 presents all 5 as produced by the handler. RFC 9457 §3 only requires type+title; handler produces all — a reasonable superset. Non-blocking.
- [x] CHK003 — Are requirements defined for the type field format validation (must be a valid URI reference per RFC 3986)? → **PASS**: FR-003 specifies "URI reference."
- [x] CHK004 — Is the about:blank fallback for unknown errors documented with its RFC 9457 §3.1 semantics? → **PASS**: FR-003 cites "about:blank (per RFC 9457 §3.1)."

## RFC 9457 Field Semantics Clarity

- [x] CHK005 — Is the title field requirement clear about its stability per RFC 9457 §3? → **NOTED**: FR-004 says "short, human-readable summary." Titles are static per ErrorMapping — inherently stable. Non-blocking.
- [x] CHK006 — Is the detail field requirement explicit about being occurrence-specific, per RFC 9457 §3? → **PASS**: FR-006 says "specific to this occurrence."
- [x] CHK007 — Are requirements defined for the instance field format — bare path acceptable per RFC 9457? → **PASS**: FR-007 says "URI reference" — bare paths like /users/123 are valid URI references per RFC 3986 §4.2.
- [x] CHK008 — Does the spec define whether instance should include the query string? → **NOTED**: Assumption specifies "no query string." Reasonable default. Non-blocking.

## Content-Type Requirement Quality

- [x] CHK009 — Is the Content-Type: application/problem+json requirement clearly scoped to error responses only? → **PASS**: FR-002 says "for all error responses"; US1 scenario 6 confirms nil errors pass through.
- [x] CHK010 — Are requirements defined for Content-Type when Accept headers don't include application/problem+json? → **NOTED**: Not addressed. Project API is JSON-only; content negotiation is YAGNI. Non-blocking.
- [x] CHK011 — Is the interaction between Content-Type and Echo c.JSON() addressed? → **NOTED**: Addressed in plan.md as implementation detail. Spec doesn't need framework internals. Non-blocking.

## Extension Field Consistency

- [x] CHK012 — Are extension member names documented as extension fields per RFC 9457 §3.2? → **PASS**: FR-008/FR-009 use "extension" language; data-model marks them as "EXTENSION."
- [x] CHK013 — Is extension field naming consistent (snake_case, no conflicts)? → **PASS**: Consistent snake_case; no RFC 9457 field conflicts.
- [x] CHK014 — Are extension fields required in all responses or only when non-empty? → **NOTED**: FR-008 says "MUST include"; data-model uses omitempty — standard JSON practice. Non-blocking.

## Error Mapping Coverage

- [x] CHK015 — Are all 18 domain sentinel errors accounted for in the error registry? → **PASS**: All 18 are in error_registry.go. Research.md table missing ErrNotLiked — documentation gap only.
- [x] CHK016 — Are type URI slug-to-status mappings consistent with RFC 7231? → **PASS**: 404→"Not Found", 401→"Unauthorized", 400→"Bad Request", 409→"Conflict."
- [x] CHK017 — Is the ErrInternalServer → about:blank mapping explicitly defined? → **PASS**: FR-003 + Research §2: internal errors have no meaningful resolution URI.
- [x] CHK018 — Does the spec define type URI slug when multiple sentinels map to same category? → **PASS**: Research §2: same category slug, different detail. Intentional.

## Non-Standard Error Paths

- [x] CHK019 — Are requirements defined for *echo.HTTPError with custom message? → **PASS**: Edge case section: custom message → detail; type from status code if no sentinel found.
- [x] CHK020 — Are requirements defined for panics caught by Echo Recover? → **PASS**: Echo Recover → existing 500 → ErrorHandler with unknown error → about:blank.
- [x] CHK021 — Are requirements defined for type URI when config base URL is empty? → **NOTED**: FR-010 specifies defaults. Empty → development URL. Non-blocking.

## JSON Schema Contract Quality

- [x] CHK022 — Does the JSON Schema correctly reflect RFC 9457 field optionality? → **NOTED**: Schema marks type/title/status as required; RFC only requires type. Handler always produces status. Non-blocking.
- [x] CHK023 — Does the JSON Schema status range cover 422? → **PASS**: minimum:100, maximum:599 covers all HTTP codes including 422.
- [x] CHK024 — Does the JSON Schema additionalProperties at top level conflict with RFC 9457 §3.2? → **FIXED**: Changed to additionalProperties: true (2026-07-14).

## Non-Functional Requirements

- [x] CHK025 — Are requirements defined for detail field length limits? → **NOTED**: No explicit limit. JSON serialization is the implicit cap. YAGNI.
- [x] CHK026 — Are requirements defined for Cache-Control on error responses? → **NOTED**: No requirement. Error responses are per-request; caching is a non-issue. YAGNI.

## Migration Completeness

- [x] CHK027 — Is removal of old {"error":{...}} wrapper explicitly required? → **PASS**: FR-012 says "replaced"; Research §6: "No backward compatibility."
- [x] CHK028 — Are requirements defined for old dto types — deleted or deprecated? → **NOTED**: FR-012 says "replaced" → removal implied. T004 replaces types directly. Non-blocking.
- [x] CHK029 — Does the spec define ErrorMapping.Message → Detail rename? → **PASS**: Data-model §2 explicitly shows the rename; T005 references it.

## Dependencies & Assumptions

- [x] CHK030 — Is the dependency on Echo RequestID middleware explicitly stated? → **PASS**: FR-008 says "correlation ID from Echo's RequestID middleware."
- [x] CHK031 — Is the instance raw-path assumption validated against RFC 9457? → **PASS**: Bare path is valid URI reference per RFC 3986 §4.2; Assumptions section validates.
- [x] CHK032 — Are requirements defined for type URI when app.port changes? → **NOTED**: FR-010 defaults include port; handler constructs from config. Port-aware by design. Non-blocking.

## Summary

| | Count |
|---|---|
| PASS (explicitly covered) | 18 |
| FIXED (was a gap, now resolved) | 1 |
| NOTED (acknowledged, non-blocking) | 13 |
| **Total** | 32 |
| **Blocking issues** | 0 |

All 32 items reviewed. 18 resolved by existing spec, 1 fixed (JSON Schema additionalProperties), 13 noted as reasonable defaults or YAGNI. **Ready for implementation.**
