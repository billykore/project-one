# Research: User Profile Edit

**Feature**: 001-user-profile-edit  
**Date**: 2026-07-08

## Research Topics

### 1. API Design: Update profile endpoint

**Decision**: Add `PUT /users/profile` (authorized) endpoint.
**Rationale**:
- Follows existing `PUT /users/password` pattern for authorized user self-mutation endpoints
- RESTful: profile is a resource belonging to the authenticated user
- Avoids `PUT /users/:username` which would be ambiguous (editing own profile vs admin editing anyone)
- The handler extracts username from JWT context (via middleware), so no URL param is needed

**Alternatives considered**:
- `PATCH /users/:username` — would need ownership check, and PATCH semantics (partial update) add unnecessary complexity when all three fields are always sent
- `PUT /users/:username/profile` — adds nesting without benefit given the endpoint is always self-referencing

### 2. Domain validation for profile updates

**Decision**: Add a focused `ValidateProfileUpdate()` method on the `User` domain entity instead of reusing the full `Validate()` (which also checks email and password).
**Rationale**:
- Profile update only touches first_name, last_name, username — not email or password
- Reusing `Validate()` would require dummy email/password values, which is a design smell
- Clean Architecture principle: domain validation should match the specific operation
- The existing `Validate()` calls are already used by `Register()` where all fields are required

**Alternatives considered**:
- Relax `Validate()` to skip password/email when empty — breaks registration validation contract
- Move validation entirely to DTO struct tags — violates Clean Architecture (domain should own validation rules)

### 3. Username uniqueness check

**Decision**: Reuse existing `GetUserByUsername()` repository method + database UNIQUE constraint as safety net.
**Rationale**:
- `userModel.Username` already has `gorm:"unique;notNull"` tag and the column has a UNIQUE constraint
- Application-level check (before UPDATE) provides a better error message than a raw DB duplicate key error
- Case-insensitive: the handler lowercases username before lookup and storage, consistent with `Register()`
- Own username check: before querying, compare lowercased input against current username to avoid false conflict

**Alternatives considered**:
- Rely on DB constraint alone — worse UX (raw error message), harder to distinguish username vs email conflicts
- Add a dedicated `IsUsernameTaken()` or `UsernameExists()` — unnecessary abstraction for a single caller; `GetUserByUsername` returning `ErrUserNotFound` or a user is sufficient

### 4. Frontend routing for edit page

**Decision**: Place edit page at `/settings/profile/edit` as a new Server Component page.
**Rationale**:
- Follows the convention of grouping user-owned settings under `/settings/`
- Server Component fetches current user data (cookie-driven auth), passes to client form component
- Avoids placing it under `/[username]/edit` which would require ownership validation in the URL

**Alternatives considered**:
- `/profile/edit` — conflicts with existing profile viewing convention
- Client-side modal instead of separate page — modals are harder to link to, worse for accessibility, and the constitution prefers page-level navigation patterns
- `/[username]/edit` — requires guarding against editing other users' profiles via URL manipulation; the `/settings/` prefix makes self-editing explicit

### 5. Form component and state management

**Decision**: Create a new `edit-profile-form.tsx` client component with local React state + Zod validation.
**Rationale**:
- Follows existing `reset-password-form.tsx` pattern: client component receives initial data as props, manages form state locally
- Zod schema for client-side validation (consistent with `changePasswordSchema` and `registerRequestBodySchema`)
- Server-side validation via `validator/v10` struct tags as the authoritative gate
- Reuses existing `InputField` component for form fields

**Alternatives considered**:
- React Hook Form — adds a dependency for a simple 3-field form; overengineered
- Server Actions — not feasible since the API proxy pattern requires HTTP calls to the Go backend
- URL state — unnecessary for a form that navigates away on success

### 6. Redirect after successful update

**Decision**: On success, the form client component calls `router.push(\`/${username}\`)` to navigate to the user's profile page. Success message passed via URL search param (`?updated=1`).
**Rationale**:
- Profile page detects the param and shows a toast/success banner
- Avoids complex state lifting across pages
- Consistent with existing redirect patterns in the codebase

**Alternatives considered**:
- Flash message in cookie — adds cookie parsing overhead for a simple message
- React Context — doesn't survive page navigation
- `router.refresh()` only — user stays on edit page, doesn't satisfy FR-011

### 7. No database migration required

**Decision**: No new migration needed. The `users` table already has `first_name`, `last_name`, `username` columns with appropriate constraints.
**Rationale**:
- Migration `000001` created `first_name` and `last_name` columns
- Migration `000006` added `username` with UNIQUE constraint
- The existing `UpdateUser` repository method performs `Save()` which updates all columns including `updated_at`

**Confirmed by**: Reviewing all migration files in `db/migrations/`.
