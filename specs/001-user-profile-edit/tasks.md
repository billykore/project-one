# Tasks: User Profile Edit

**Input**: Design documents from `/specs/001-user-profile-edit/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/, quickstart.md

**Tests**: Use `make test` for backend and `npm test` for frontend to verify existing coverage. See T028 for the new `TestUpdateProfile` unit test.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Backend**: Go source under `internal/`, route registration in `cmd/main.go`
- **Frontend**: Next.js source under `web/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm project structure and tooling readiness — no new scaffolding needed

- [X] T001 Verify `make check` passes on clean main branch to establish baseline
- [X] T002 Verify `cd web && npm test` passes on clean main branch to establish baseline

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Domain entity, port interfaces, and DTOs that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T003 [P] Add `ValidateProfileUpdate()` method to User domain entity in `internal/core/domain/user.go` — validates first_name (required, 3-100 chars, trimmed), last_name (required, 3-100 chars, trimmed), username (required, 3-30 chars, matches `^[a-zA-Z0-9_]+$`, case-insensitive via lowercasing). Must NOT validate email or password.
- [X] T004 [P] Add `UpdateProfile(ctx, username string, user *domain.User) error` method to `UserUseCase` interface in `internal/core/ports/user.go`
- [X] T005 [P] Add `UpdateProfile(ctx context.Context, user *domain.User) error` method to `UserRepository` interface in `internal/core/ports/user.go` — distinct from existing `UpdateUser`; must handle username cascade within a transaction
- [X] T006 [P] Add `UpdateProfileRequest` and `UpdateProfileResponse` DTOs in `internal/api/dto/user_dto.go`:
  - `UpdateProfileRequest`: FirstName (`validate:"required,min=3,max=100"`), LastName (`validate:"required,min=3,max=100"`), Username (`validate:"required,min=3,max=30"`)
  - `UpdateProfileResponse`: Message string, Username string
- [X] T007 [P] Add Zod validation schema `editProfileSchema` in `web/lib/types/profile.types.ts`:
  - first_name: trimmed, min 3, max 100
  - last_name: trimmed, min 3, max 100
  - username: trimmed, min 3, max 30, regex `^[a-zA-Z0-9_]+$`
  - Export `EditProfileFormData` type from schema

**Checkpoint**: Foundation ready — all interfaces, domain validation, and DTOs defined. User story implementation can now begin.

---

## Phase 3: User Story 1 - Authenticated User Edits Profile (Priority: P1) 🎯 MVP

**Goal**: An authenticated user can navigate from their profile page to a dedicated edit form, modify their first name, last name, and/or username, submit the changes, and see the updated data on their profile page with a success confirmation.

**Independent Test**: Log in as a user, navigate to profile page, click "Edit Profile", modify first name, submit — verify profile page shows updated name with success banner.

### Implementation for User Story 1

#### Backend — Use Case & Repository

- [X] T008 [US1] Implement `UpdateProfile()` method in `internal/core/usecase/user_usecase.go`:
  - Trim whitespace from all fields, lowercase username
  - Call `user.ValidateProfileUpdate()`
  - If username changed: call `GetUserByUsername` to check uniqueness (handle `ErrUserNotFound` as OK; any other user → `ErrUsernameAlreadyTaken`)
  - Call `userRepo.UpdateProfile(ctx, user)` (transactional cascade)
  - Return updated username for response
- [X] T009 [US1] Implement `UpdateProfile()` method in `internal/adapters/repository/user_repository.go`:
  - Use `gorm.DB.Transaction` to wrap all operations
  - Update the `users` row (first_name, last_name, username, updated_at) by user ID
  - If username changed: cascade UPDATE to `follows.follower_username`, `follows.followed_username`, `posts.username`, `user_tokens.username`, `comments.username`, `notifications.actor_username`, `post_likes.username`
  - Handle `gorm.ErrDuplicatedKey` on username → return `domain.ErrUsernameAlreadyTaken`
- [X] T028 [US1] Create `TestUpdateProfile` unit test in `internal/core/usecase/user_usecase_test.go` using GoMock-generated mocks for `UserRepository` and `Hasher` interfaces. Cover: (a) happy path — valid profile update calls `UpdateProfile` and returns updated username, (b) username changed — new username passes uniqueness check, cascade triggered, (c) username unchanged — no uniqueness lookup, `UpdateProfile` still called, (d) username conflict — `GetUserByUsername` returns another user, `ErrUsernameAlreadyTaken` returned, (e) validation failure — empty first name returns `ErrValidationFailed` with descriptive message. Run `make mock` first if mocks are stale.

#### Backend — Handler & Route

- [X] T010 [US1] Implement `HandleUpdateProfile()` handler in `internal/api/handler/user_handler.go`:
  - Extract `username` from context (`c.Get("username")`)
  - Bind and validate `UpdateProfileRequest` DTO
  - Trim whitespace from all fields in handler before passing to use case
  - Call `userUseCase.UpdateProfile(ctx, username, user)`
  - Map errors: `ErrValidationFailed` → 400, `ErrUsernameAlreadyTaken` → 400, others → 500
  - Return 200 with `UpdateProfileResponse`
- [X] T011 [US1] Register `PUT /users/profile` route in `cmd/main.go`:
  - Add to `usersAuth` group (already behind `middleware.Authorize(tokenSvc)`)
  - Wire to `userHdl.HandleUpdateProfile`

#### Backend — Swagger

- [X] T012 [US1] Add Swagger annotations to `HandleUpdateProfile` handler in `internal/api/handler/user_handler.go` and run `make docs` to regenerate `api/swagger/`

#### Frontend — Edit Profile Button

- [X] T013 [US1] Add "Edit Profile" button to `web/components/profile/user-profile-view.tsx`:
  - Render only when `isOwner === true` (already computed by `useProfileController`)
  - Use `<Link href="/settings/profile/edit">` with existing indigo button styling
  - Place in the sidebar User Meta Card, below the followers/following stats section

#### Frontend — Edit Profile Form Component

- [X] T014 [P] [US1] Create `web/components/profile/edit-profile-form.tsx` ("use client"):
  - Props: `initialFirstName`, `initialLastName`, `initialUsername` (strings)
  - Local state for form fields, errors record, submitting boolean, serverError string
  - Render three `InputField` components (First Name, Last Name, Username) from `web/components/ui/input.tsx`
  - Client-side Zod validation on submit using `editProfileSchema` from `web/lib/types/profile.types.ts`
  - On validation pass: PUT to `/api/users/profile` with JSON body
  - On 200: call `router.push(\`/${username}?updated=1\`)` (use `useRouter` from next/navigation)
  - On 400: parse error message; if "already taken" → show serverError banner; else show field-level errors
  - On network error: show serverError banner with "Unable to connect. Please try again."
  - Loading state: disable submit button, show spinner text ("Saving...")
  - Cancel button: `<Link href="\`/${initialUsername}\`">` or `router.back()`
  - Follow glassmorphism card pattern (`rounded-2xl bg-white/80 backdrop-blur border border-zinc-200 dark:border-zinc-800 p-6 shadow-sm`)
  - Dark mode support via Tailwind `dark:` variants
  - Accessibility: `aria-invalid` on error fields, `aria-describedby` for error messages

#### Frontend — Edit Profile Page

- [X] T015 [US1] Create `web/app/settings/profile/edit/page.tsx` (Server Component):
  - Read `username` cookie from `cookies()` (from `next/headers`)
  - If no username cookie → `redirect("/login")`
  - Fetch user data via `serverFetch(\`/api/users/${username}\`)` with `handleApiResponse<UserProfile>`
  - On 401 → `redirect("/login")`
  - On error → render error state with retry
  - While loading → render `PageSkeletonLayout` from `web/components/ui/page-skeleton.tsx`
  - On success → render `<EditProfileForm>` with `initialFirstName`, `initialLastName`, `initialUsername` extracted from user data
  - Parse `name` field (format "FirstName LastName") to extract firstName and lastName

#### Frontend — Success Banner on Profile Page

- [X] T016 [US1] Add success banner to `web/components/profile/user-profile-view.tsx`:
  - Read `?updated=1` search param via `useSearchParams()` from `next/navigation`
  - If param present, show a green success banner: "✅ Profile updated successfully" above the profile header
  - Banner should auto-dismiss after 5 seconds (use `useEffect` + `setTimeout`) or be manually dismissible via ✕ button
  - Wrap in `Suspense` if using `useSearchParams`

**Checkpoint**: User Story 1 is fully functional — end-to-end profile edit flow works: navigate → edit → save → see updated profile.

---

## Phase 4: User Story 2 - Username Uniqueness Enforcement (Priority: P2)

**Goal**: When a user attempts to change their username to one already taken by another user, the system rejects with a clear error and preserves their form data.

**Independent Test**: User A edits profile, enters User B's username, submits — error "This username is already taken" appears, first name and last name are preserved in form.

### Implementation for User Story 2

- [X] T017 [US2] Verify username uniqueness logic in `internal/core/usecase/user_usecase.go` `UpdateProfile()`:
  - Confirm `GetUserByUsername` is called only when username actually changed (compare lowercased old vs new)
  - Confirm `ErrUserNotFound` from `GetUserByUsername` is treated as "available" (not a conflict)
  - Confirm any returned user (non-ErrUserNotFound) triggers `ErrUsernameAlreadyTaken`
  - Ensure the `users` table UNIQUE constraint catches races (application check + DB safety net)
- [X] T018 [US2] Verify username uniqueness error display in `web/components/profile/edit-profile-form.tsx`:
  - Parse 400 response body for "already taken" substring
  - Display as `serverError` banner (not field-level on username input)
  - Confirm form state is preserved (firstName, lastName, attempted username remain in inputs)
  - Username input should NOT be cleared on uniqueness error

**Checkpoint**: Username uniqueness is enforced with proper error UX. User can retry with a different username without losing other edits.

---

## Phase 5: User Story 3 - Unauthenticated Access Guard (Priority: P3)

**Goal**: Unauthenticated users who navigate directly to `/settings/profile/edit` are redirected to login. After logging in, they are returned to the edit page.

**Independent Test**: Log out, navigate to `/settings/profile/edit` — verify redirect to `/login`. Log in — verify redirect back to `/settings/profile/edit`.

### Implementation for User Story 3

- [X] T019 [US3] Verify `web/proxy.ts` auth guard covers `/settings/profile/edit`:
  - `/settings` is not in `PUBLIC_ROUTES` set, but is it matched by the matcher?
  - The matcher `/((?!_next/static|_next/image|favicon.ico|.*\\.png$).*)` should catch `/settings/profile/edit`
  - Manually test: log out, visit `/settings/profile/edit`, confirm redirect to `/login`
- [X] T020 [US3] Verify login redirect preserves return URL for edit page:
  - Confirm the login page reads `?redirect=` or `returnUrl` param and redirects post-login
  - If login page does not currently support return URLs, add support: read search param, store in hidden field or state, redirect on success

**Checkpoint**: Unauthenticated users cannot access the edit page. Login redirect preserves intent.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Testing, documentation, and final validation

- [X] T021 [P] Run `make mock` to regenerate GoMock mocks after interface changes
- [X] T022 [P] Run `make check` (docs + vet + lint + test) and fix any issues
- [X] T023 [P] Run `cd web && npm test` and fix any test failures
- [X] T024 Run `make docs` to ensure Swagger docs are up to date with new endpoint
- [X] T025 Execute quickstart.md validation scenarios end-to-end (happy path, uniqueness, empty fields, unauthenticated, cascade)
- [X] T026 [P] Verify responsive layout: test edit profile form on mobile viewport (375px width)
- [X] T027 [P] Verify dark mode: toggle dark mode, confirm form fields, banners, and button styling

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — verify baseline immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational phase completion
- **User Story 2 (Phase 4)**: Depends on User Story 1 completion (validates/extends existing use case + form)
- **User Story 3 (Phase 5)**: Depends on User Story 1 completion (needs the edit page route to exist)
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2). No dependencies on other stories.
- **User Story 2 (P2)**: Can start after US1. Extends the same use case and form component — verifies error path of US1's implementation.
- **User Story 3 (P3)**: Can start after US1. Only touches proxy.ts and login page — minimal changes.

### Within Each User Story

- Backend use case → backend repository (T008 depends on T009 conceptually, but can be done together)
- Backend handler → route registration (T010 → T011, sequential)
- Frontend components can be built in parallel with backend once interfaces/DTOs are defined
- Backend handler + route registration should be done before frontend form submits to the endpoint

### Parallel Opportunities

- Phase 2: T003, T004, T005, T006, T007 can all run in parallel (different files)
- Phase 3: T014 (frontend form) can be built in parallel with T008-T012 (backend)
- Phase 6: T021, T022, T023, T026, T027 can all run in parallel

---

## Parallel Example: User Story 1

```text
          ┌── T008 (UseCase) ──┐
T003-T007 ─┤                   ├── T010 (Handler) ── T011 (Route) ── T012 (Swagger)
(Phase 2)  └── T014 (Form)  ──┬── T013 (Button) ── T015 (Page) ──── T016 (Banner)
                              │
                              └── (Frontend can proceed in parallel with backend)
```

---

## Implementation Strategy

### MVP Scope (Deliverable 1)

**Phase 1 + 2 + 3**: Foundation + User Story 1

This delivers the complete profile edit flow:
- Backend: `PUT /users/profile` endpoint with validation and username cascade
- Frontend: Edit Profile button → Edit Profile page → Edit Profile form → Success redirect

Users can edit their first name, last name, and username. ✓

### Incremental Delivery

1. **Deliverable 1 (MVP)**: User Story 1 — Complete edit profile flow (Phases 1-3)
2. **Deliverable 2**: User Story 2 — Username uniqueness enforcement (Phase 4)  
3. **Deliverable 3**: User Story 3 — Auth guard for unauthenticated access (Phase 5)
4. **Deliverable 4**: Polish — Testing, documentation, responsive/dark mode validation (Phase 6)
