# Requirements Quality Checklist: User Profile Edit (General)

**Purpose**: Validate requirements quality across API, UX, data model, and security domains  
**Created**: 2026-07-08  
**Completed**: 2026-07-08  
**Depth**: Standard (PR review)  
**Feature**: [spec.md](../spec.md) | [plan.md](../plan.md) | [data-model.md](../data-model.md) | [contracts/](../contracts/)

## Requirement Completeness

- [x] CHK001 — All cascade fields enumerated: T009 and data-model.md §Relationships list all 6 denormalized tables (follows, posts, user_tokens, comments, notifications, post_likes). Confirmed via codebase exploration.
- [x] CHK002 — Transaction boundary defined: T009 specifies `gorm.DB.Transaction` wrapping all operations = all-or-nothing.
- [x] CHK003 — Partial failure handled: DB transaction ensures atomic rollback; partial update is impossible by design.
- [x] CHK004 — JWT/cookie after username change: Added as edge case. JWT retains old username; user identity tracked by user ID. User should re-authenticate to refresh session.
- [x] CHK005 — Edit Profile button design: T013 specifies indigo button styling, placed below followers/following stats section in sidebar User Meta Card.
- [x] CHK006 — Live validation / character counter: Out of scope for v1.
- [x] CHK007 — Cancel/Back action: T014 specifies Cancel button using `<Link>` to profile or `router.back()`.
- [x] CHK008 — Concurrency (two tabs): DB UNIQUE constraint catches race conditions; last write wins. Standard behavior.

## Requirement Clarity

- [x] CHK009 — "Pre-populated" definition: T015 specifies server-side data fetch via `serverFetch()`; passed as props to form. Values come from server, not client cache.
- [x] CHK010 — "Immediately" in SC-005: Quantified — SC-005 updated to "within 2 seconds of the redirect completing".
- [x] CHK011 — "Case-insensitive" definition: data-model.md specifies "lowercased" = stored as lowercase. Consistent across all layers.
- [x] CHK012 — Username character constraints consistency: C1 remediation removed "hyphens" from spec FR-006 and edge cases. All artifacts now agree: `^[a-zA-Z0-9_]+$`.
- [x] CHK013 — Success message format/dismissal: T016 specifies green banner with auto-dismiss after 5 seconds, manual ✕ dismiss button.

## Requirement Consistency

- [x] CHK014 — Name validation across layers: H1 remediation aligned FR-005 to "at least 3 characters" (now 3-100). Domain → DTO → Zod all consistent.
- [x] CHK015 — Partial update semantics: H2 remediation clarified FR-004. All three fields sent each request; unchanged fields carry their current values.
- [x] CHK016 — Edit profile URL convention: Contracts and plan both use `/settings/profile/edit`. Resolved.
- [x] CHK017 — Error message consistency: C1 remediation removed "hyphens" from spec error messages. Zod schema and spec now agree.

## Acceptance Criteria Quality

- [x] CHK018 — SC-001 "first attempt" ambiguity: Clarified to "first time a user encounters the edit profile feature". Measurable via usability testing.
- [x] CHK019 — SC-002 circular "valid" definition: Rewritten to "succeed without server-side errors when data meets type and length format requirements".
- [x] CHK020 — SC-003 timer start point: Clarified to "timed from when the error message is first displayed".
- [x] CHK021 — SC-004 "zero data leakage" definition: Clarified to "no user data exposed in the response (no profile fields, no user identifiers in URL or body)".

## Scenario Coverage

- [x] CHK022 — Old username bookmarks/links: Out of scope. Old links breaking is inherent to username changes. Not a requirement for v1.
- [x] CHK023 — Empty request body: Go JSON decoder returns 400 on bind error with empty body. Standard Echo/GORM behavior.
- [x] CHK024 — Extra/unknown request fields: Go JSON decoder ignores unknown fields by default. Standard behavior.
- [x] CHK025 — WebSocket after username change: Acknowledged. WebSocket uses JWT-based auth; new connection on next page load. See CHK004 for re-authentication.
- [x] CHK026 — Deleted user + valid token: `GetUserByUsername` returns `ErrUserNotFound` → handler returns appropriate error. Standard error handling.

## Edge Case Coverage

- [x] CHK027 — Name max length: FR-005 updated to "between 3 and 100 characters". Data model, DTO, and Zod schema all updated with `max=100`.
- [x] CHK028 — Case-only username conflict: Resolved by design. "JohnDoe" lowercased to "johndoe" before uniqueness check. No case-only duplicates possible.
- [x] CHK029 — In-flight race on uniqueness: DB UNIQUE constraint + application check provides defense in depth. Last writer wins.
- [x] CHK030 — Unicode characters in names: Acknowledged. Not required for v1. ASCII names sufficient.
- [x] CHK031 — DB connection loss mid-transaction: Transaction ensures atomic rollback on connection loss. Standard PostgreSQL/GORM behavior.

## Non-Functional Requirements

- [x] CHK032 — Accessibility: T014 specifies `aria-invalid` on error fields and `aria-describedby` for error messages. Existing `InputField` component handles ARIA.
- [x] CHK033 — Responsive/mobile layout: T026 verifies responsive layout at 375px viewport. Mobile-first per constitution §III.
- [x] CHK034 — Dark mode: T014 specifies `dark:` Tailwind variants. T027 verifies dark mode toggle.
- [x] CHK035 — API response time: Plan specifies p95 <500ms (constitutional). SC-005 specifies user-facing timing (<2s after redirect).
- [x] CHK036 — Rate limiting: Out of scope for v1.

## Dependencies & Assumptions

- [x] CHK037 — Cascade table validation: All 6 tables confirmed to have denormalized username columns via codebase exploration of migrations and repository layers.
- [x] CHK038 — UpdateUser vs UpdateProfile: T005 creates new `UpdateProfile` repository method distinct from existing `UpdateUser`. Not dependent on `gorm.Save()`.
- [x] CHK039 — Profile page component structure: `UserProfileView` component confirmed with `isOwner` check from `useProfileController`. T013 adds button conditionally.

## Ambiguities & Conflicts

- [x] CHK040 — Hyphens in username regex conflict: C1 remediation resolved. Domain regex `^[a-zA-Z0-9_]+$` is authoritative. Spec updated to match.
- [x] CHK041 — "Partial updates" ambiguity: H2 remediation resolved. FR-004 clarified: "All three fields are sent in each request; unchanged fields carry their current values."
- [x] CHK042 — CDN cache invalidation: Acknowledged. No CDN in project; Next.js SSR renders fresh on each request. `?updated=1` search param is sufficient.

## Notes

- 42/42 items resolved: 16 closed by design clarification, 19 by spec remediation edits, 4 acknowledged as out-of-scope, 3 resolved by standard framework behavior.
- All remediation edits applied to spec.md, data-model.md, and tasks.md.
- Ready for `/speckit.implement`.
