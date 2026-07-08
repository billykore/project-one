# Feature Specification: User Profile Edit

**Feature Branch**: `001-user-profile-edit`

**Created**: 2026-07-08

**Status**: Draft

**Input**: User description: "Develop a feature for user to be able to edit their profile. User must be logged in to be able to changes their profile. Lets redirect the user from the profile page to the new edit profile page for user to change their data. Data that user can changes: First name, Last name, Username."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Authenticated User Edits Profile (Priority: P1)

An authenticated user visits their own profile page, clicks an "Edit Profile" action, and is redirected to a dedicated edit profile page. The form is pre-populated with their current first name, last name, and username. The user modifies one or more fields and submits. The system validates the input, updates the profile, and redirects the user back to their profile page where the updated data is displayed. A success confirmation message is shown.

**Why this priority**: This is the core feature — without it, profile editing does not exist. It delivers the primary user value of allowing users to manage their own identity on the platform.

**Independent Test**: Can be fully tested by logging in as a user, navigating to the edit profile page, modifying first name/last name/username, submitting the form, and verifying the profile page reflects the changes. Delivers complete self-service profile management.

**Acceptance Scenarios**:

1. **Given** a logged-in user is viewing their own profile page, **When** they click "Edit Profile", **Then** they are redirected to the edit profile page with their current first name, last name, and username pre-filled in the form.
2. **Given** a logged-in user is on the edit profile page with valid changes entered, **When** they submit the form, **Then** their profile is updated, they are redirected to their profile page, the updated data is displayed, and a success message is shown.
3. **Given** a logged-in user is on the edit profile page, **When** they submit the form with no changes to any field, **Then** the system accepts the submission without error and redirects back to the profile page (idempotent save).
4. **Given** a logged-in user is on the edit profile page, **When** they modify only their first name and submit, **Then** only the first name is updated and last name/username remain unchanged.

---

### User Story 2 - Username Uniqueness Enforcement (Priority: P2)

When a user attempts to change their username to one that is already taken by another user, the system rejects the submission and displays a clear error message asking them to choose a different username. The form retains all other entered data so the user does not lose their work.

**Why this priority**: Username uniqueness is critical for platform integrity (mentions, follows, routing). Without this, duplicate usernames would break core social features. It is P2 because it handles a validation edge case rather than the happy path.

**Independent Test**: Can be tested by having User A change their username to one already used by User B, submitting, and verifying the error message appears and form data is preserved.

**Acceptance Scenarios**:

1. **Given** a logged-in user is on the edit profile page, **When** they enter a username that already belongs to another user and submit, **Then** the system rejects the submission and displays an error: "This username is already taken. Please choose another."
2. **Given** a logged-in user is on the edit profile page, **When** they re-enter their current username (no change) and submit, **Then** the system accepts it without a uniqueness conflict (own username is allowed).

---

### User Story 3 - Unauthenticated Access Guard (Priority: P3)

A user who is not logged in attempts to access the edit profile page directly via URL. The system denies access and redirects them to the login page. After successful login, they are returned to the edit profile page they originally attempted to reach.

**Why this priority**: Security and proper access control are essential, but this is a guard rail for a secondary flow (direct URL access). The primary flow already requires being on one's own profile page, which requires authentication.

**Independent Test**: Can be tested by logging out, navigating directly to the edit profile URL, and confirming redirection to login. After logging in, verify the user lands on the edit profile page.

**Acceptance Scenarios**:

1. **Given** an unauthenticated user, **When** they navigate directly to the edit profile URL, **Then** they are redirected to the login page.
2. **Given** an unauthenticated user was redirected to login from the edit profile URL, **When** they successfully log in, **Then** they are redirected back to the edit profile page.

---

### Edge Cases

- What happens when a user submits an empty first name or last name? → System rejects with a validation error indicating which field is required.
- What happens when the username is too short (e.g., fewer than 3 characters) or too long (e.g., more than 30 characters)? → System rejects with appropriate length validation messaging.
- What happens when the username contains invalid characters (spaces, special symbols)? → System rejects with a format validation message (e.g., "Username may only contain letters, numbers, and underscores").
- What happens when the user's authentication token expires mid-edit? → System rejects the save attempt and redirects to login, preserving the intent to return.
- What happens when a network error occurs during submission? → System displays a user-friendly error message and allows retry without data loss.
- What happens when two users attempt to claim the same username simultaneously? → Only one succeeds; the other receives a uniqueness conflict error.
- What happens when leading/trailing whitespace is entered in fields? → System trims whitespace before validation and storage.
- What happens when a user changes their username while logged in? → The existing JWT and username cookie retain the old username until the user logs out and logs back in. Profile edits succeed based on the authenticated user identity (user ID), not the username in the token. The user should re-authenticate to refresh their session with the new username.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide an "Edit Profile" action accessible from the authenticated user's own profile page.
- **FR-002**: The system MUST redirect the authenticated user from their profile page to a dedicated edit profile page when the "Edit Profile" action is activated.
- **FR-003**: The edit profile page MUST display a form pre-populated with the user's current first name, last name, and username.
- **FR-004**: The system MUST allow the user to modify their first name, last name, and username. All three fields are sent in each request; unchanged fields carry their current values.
- **FR-005**: The system MUST validate that first name and last name are between 3 and 100 characters after trimming whitespace.
- **FR-006**: The system MUST validate that the username meets length requirements (3–30 characters) and character constraints (alphanumeric and underscores).
- **FR-007**: The system MUST enforce username uniqueness across all users (case-insensitive). Attempting to take another user's username MUST be rejected.
- **FR-008**: The system MUST allow a user to re-submit their current username without triggering a uniqueness conflict.
- **FR-009**: The system MUST redirect unauthenticated users who attempt to access the edit profile page to the login page.
- **FR-010**: After successful login from a redirect, the system MUST return the user to the edit profile page.
- **FR-011**: Upon successful profile update, the system MUST redirect the user to their profile page and display a success confirmation message.
- **FR-012**: Upon validation failure, the system MUST redisplay the edit form with the user's entered data preserved and show field-level error messages.
- **FR-013**: The system MUST trim leading and trailing whitespace from all text fields before validation and storage.
- **FR-014**: The system MUST only allow a user to edit their own profile. Attempts to edit another user's profile via direct API calls MUST be rejected with an authorization error.
- **FR-015**: The "Edit Profile" action MUST only be visible on the user's own profile page, not when viewing another user's profile.

### Key Entities

- **User Profile**: Represents a user's publicly visible identity on the platform. Key attributes include first name, last name, and username. Each profile is owned by exactly one authenticated user. The username serves as a unique, human-readable identifier across the platform and is used in follows, mentions, and profile URLs.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users encountering the edit profile feature for the first time can navigate from their profile page to the edit form and submit changes in under 30 seconds.
- **SC-002**: 95% of profile edit submissions succeed without server-side errors when the submitted data meets all type and length format requirements.
- **SC-003**: Invalid submissions (duplicate username, empty fields) produce clear, field-specific error messages. Users can read the error, correct their input, and resubmit in under 10 seconds (timed from when the error message is first displayed).
- **SC-004**: Unauthenticated users who navigate to the edit profile page are redirected to login with no user data exposed in the response (no profile fields, no user identifiers in URL or body).
- **SC-005**: Profile updates are reflected on the profile page within 2 seconds of the redirect completing — users never see stale data after a successful edit.

## Assumptions

- Users are already authenticated via the existing JWT-based authentication system. This feature does not introduce new authentication mechanisms.
- The existing user profile page (`web/app/[username]/page.tsx` or equivalent) already exists and can have an "Edit Profile" action added to it.
- The existing backend user model and API infrastructure (Echo, GORM, validator/v10) will be extended to support profile updates. No new database tables are required; the `users` table already has columns for first name, last name, and username (as evidenced by migration `000006_add_username_to_users.up.sql`).
- Username uniqueness is enforced at the database level (unique constraint/index) in addition to application-level validation.
- The frontend will follow the existing glassmorphism design system and use existing UI components (shadcn/ui + Radix primitives) as required by the project constitution.
- The edit profile page will be a client-side rendered page (`"use client"`) due to form interactivity requirements, with appropriate justification per the constitution's performance principle.
- The edit profile URL will follow the existing routing convention (e.g., `/settings/profile/edit` or `/profile/edit`).
- Guest UX guards (disabled interactions with tooltips) already exist for unauthenticated users on profile pages and will be extended to cover the edit profile action.
