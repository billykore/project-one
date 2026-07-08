# Implementation Plan: User Profile Edit

**Branch**: `001-user-profile-edit` | **Date**: 2026-07-08 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-user-profile-edit/spec.md`

## Summary

Add a profile editing capability allowing authenticated users to modify their first name, last name, and username. Extend the existing Go backend with a new `PUT /users/profile` endpoint that validates input, enforces username uniqueness, and updates the user record. On the frontend, add an "Edit Profile" button to the profile page that redirects to a new `/settings/profile/edit` page with a pre-populated form. After successful submission, redirect back to the profile page with a success confirmation.

## Technical Context

**Language/Version**: Go 1.26+ (backend), TypeScript (frontend)

**Primary Dependencies**: Echo (HTTP framework), GORM (ORM), validator/v10 (validation), Zerolog (logging), Next.js 16 (App Router), React 19, Tailwind CSS 4, Zod (validation), shadcn/ui + Radix primitives (UI components)

**Storage**: PostgreSQL (via GORM), existing `users` table with columns `id`, `email`, `username`, `password`, `first_name`, `last_name`, `created_at`, `updated_at`

**Testing**: Go: testify + GoMock (`make mock`), frontend: Vitest (`npm test`)

**Target Platform**: Linux server (backend), modern web browsers (frontend)

**Project Type**: Web application (Go API backend + Next.js frontend)

**Performance Goals**: p95 <500ms for profile write endpoint, no N+1 queries (single-row UPDATE)

**Constraints**: Must follow Clean Architecture layer separation; no business logic in handlers; JWT auth required for mutation; guest UX with disabled interactions and tooltips

**Scale/Scope**: Single-user profile edit (one user editing their own profile at a time); username uniqueness enforced across all users

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Check (Phase 0)

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Code Quality & Clean Architecture** | ✅ PASS | New `UpdateProfile` use case in `usecase/`, new handler method in existing `UserHandler`, no new layers needed. |
| **II. Testing Standards** | ✅ PASS | UseCase unit test with GoMock mocks. Frontend component test with Vitest. |
| **III. User Experience Consistency** | ✅ PASS | Reuses existing `InputField`, glassmorphism patterns, Zod validation, loading/error/success states. |
| **IV. Performance Requirements** | ✅ PASS | Single-user profile mutation. No pagination concerns. |
| **Security** | ✅ PASS | JWT auth, input validation, ownership enforcement. |
| **Development Workflow** | ✅ PASS | Conventional Commits, Swagger docs, pre-merge gates. |

### Post-Design Re-evaluation (Phase 1)

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Code Quality & Clean Architecture** | ✅ PASS | `UpdateProfile` use case orchestrates: validate domain entity → check username uniqueness → update user + cascade username to denormalized columns (follows, posts, user_tokens, comments, notifications, post_likes). All within a transaction at the repository layer. Handler delegates all logic to use case (no business logic leak). |
| **II. Testing Standards** | ✅ PASS | UseCase tests must mock `UserRepository.UpdateProfile()` (transactional) and verify username cascade logic. Frontend form tests must cover validation states, loading, and error/success flows. |
| **III. User Experience Consistency** | ✅ PASS | New `/settings/profile/edit` route. Edit button on profile page (owner-only). Success redirect with `?updated=1` param triggers banner on profile page. All states covered (loading via skeleton, empty N/A, error banners, field-level errors). |
| **IV. Performance Requirements** | ✅ PASS | Username cascade requires `UPDATE` statements on multiple tables within a single DB transaction. All columns involved have indexes (username is UNIQUE on users, indexed on follows, user_tokens). Expected <100ms for the full transaction under normal load. Well within the <500ms write budget. |
| **Security: Input Validation** | ✅ PASS | Three-layer validation: (1) Zod on client → (2) validator/v10 struct tags on DTO → (3) domain validation in `ValidateProfileUpdate()`. DB UNIQUE constraint as final safety net. |
| **Security: Authorization** | ✅ PASS | `PUT /users/profile` extracts username from JWT via middleware — no URL param to spoof. Own-profile-only by design. |
| **Security: Password Storage** | ✅ N/A | Unchanged. |
| **Development Workflow** | ✅ PASS | Unchanged. One new route, one new page, one new component, modifications to existing files. All within established patterns. | |

## Project Structure

### Documentation (this feature)

```text
specs/001-user-profile-edit/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (speckit.tasks)
```

### Source Code (repository root)

```text
# Backend (Go)
internal/
├── core/
│   ├── domain/
│   │   └── user.go                    # (MODIFY) Add UpdateProfile validation helper
│   ├── ports/
│   │   └── user.go                    # (MODIFY) Add UpdateProfile to UserUseCase interface
│   └── usecase/
│       └── user_usecase.go            # (MODIFY) Implement UpdateProfile method
├── api/
│   ├── dto/
│   │   └── user_dto.go                # (MODIFY) Add UpdateProfileRequest/Response DTOs
│   └── handler/
│       └── user_handler.go            # (MODIFY) Add HandleUpdateProfile handler
└── adapters/
    └── repository/
        └── user_repository.go         # (REUSE) UpdateUser already exists

cmd/
└── main.go                            # (MODIFY) Register PUT /users/profile route

# Frontend (Next.js)
web/
├── app/
│   └── settings/
│       └── profile/
│           └── edit/
│               └── page.tsx           # (NEW) Edit profile page (Server Component)
├── components/
│   └── profile/
│       ├── user-profile-view.tsx      # (MODIFY) Add "Edit Profile" button for owner
│       └── edit-profile-form.tsx      # (NEW) Edit profile form client component
├── hooks/
│   └── use-profile-controller.ts      # (MODIFY) Optional: add edit navigation helper
└── lib/
    └── types/
        └── profile.types.ts           # (MODIFY) Add edit profile types + Zod schema
```

**Structure Decision**: Web application (Option 2). Backend extends existing Clean Architecture layers — no new layers, only new methods on existing interfaces/structs. Frontend follows existing page pattern (Server Component for data fetch, Client Component for interactive form).

## Complexity Tracking

> No violations to justify. All constitution principles are satisfied with the proposed design.

