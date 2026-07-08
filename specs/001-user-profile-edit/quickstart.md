# Quickstart: User Profile Edit

**Feature**: 001-user-profile-edit  
**Date**: 2026-07-08

Validation guide for manually testing the profile edit feature end-to-end.

## Prerequisites

1. Backend running: `make run` (Go API on `:8080`)
2. Frontend running: `cd web && npm run dev` (Next.js on `:3000`)
3. PostgreSQL with migrations applied: `make migrate-up`
4. At least two registered users (for username uniqueness testing)

## Setup Commands

```bash
# Terminal 1: Start backend
make run

# Terminal 2: Start frontend
cd web && npm run dev
```

## Validation Scenarios

### 1. Happy Path: Edit Profile

| Step | Action | Expected |
|------|--------|----------|
| 1 | Log in as user A | Redirected to home |
| 2 | Navigate to user A's profile (`/[username]`) | Profile page loads with user details |
| 3 | Click "Edit Profile" button | Redirected to `/settings/profile/edit` |
| 4 | Verify form is pre-populated | First Name, Last Name, Username match current values |
| 5 | Change First Name to "UpdatedName" | Field shows new value |
| 6 | Click "Save Changes" | Loading spinner, then redirected to profile page |
| 7 | Verify profile page | First Name shows "UpdatedName", success banner visible |

### 2. Username Uniqueness

| Step | Action | Expected |
|------|--------|----------|
| 1 | Log in as user A | - |
| 2 | Navigate to `/settings/profile/edit` | Edit form loaded |
| 3 | Change username to user B's username | - |
| 4 | Submit | Error: "This username is already taken. Please choose another." |
| 5 | Verify form data preserved | First Name, Last Name unchanged; username input shows attempted value |

### 3. Keeping Own Username (No Conflict)

| Step | Action | Expected |
|------|--------|----------|
| 1 | Log in as user A | - |
| 2 | Navigate to `/settings/profile/edit` | - |
| 3 | Change only First Name, keep same username | - |
| 4 | Submit | Success, profile updated |

### 4. Validation: Empty Fields

| Step | Action | Expected |
|------|--------|----------|
| 1 | Log in, navigate to `/settings/profile/edit` | - |
| 2 | Clear First Name field | - |
| 3 | Submit | Field-level error: "First name must be at least 3 characters" |
| 4 | Clear Last Name field | - |
| 5 | Submit | Also shows error for Last Name |
| 6 | Enter invalid username (e.g., "ab" or "user name!") | - |
| 7 | Submit | Error explaining constraints |

### 5. Unauthenticated Access

| Step | Action | Expected |
|------|--------|----------|
| 1 | Log out (or use incognito window) | - |
| 2 | Navigate directly to `/settings/profile/edit` | Redirected to `/login` |
| 3 | Log in | Redirected back to `/settings/profile/edit` |

### 6. Username Cascade Verification

| Step | Action | Expected |
|------|--------|----------|
| 1 | Log in as user A who has posts and followers | - |
| 2 | Change username via edit profile | Success |
| 3 | Visit profile at new username URL | Profile loads with old posts visible under new username |
| 4 | Check that followers/following still show correct username | No broken references |

## API-Only Testing

```bash
# Login to get token
curl -s -c cookies.txt -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Update profile
curl -s -b cookies.txt -X PUT http://localhost:8080/users/profile \
  -H "Content-Type: application/json" \
  -d '{"first_name":"NewFirst","last_name":"NewLast","username":"newusername"}'

# Expected: {"message":"Profile updated successfully","username":"newusername"}

# Verify with duplicate username
curl -s -b cookies.txt -X PUT http://localhost:8080/users/profile \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Test","last_name":"Test","username":"existing_other_user"}'

# Expected: {"error":"Username is already taken"}

# Unauthorized (no cookie)
curl -s -X PUT http://localhost:8080/users/profile \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Test","last_name":"Test","username":"test"}'

# Expected: {"error":"Unauthorized"}
```

## Backend Tests

```bash
# Run all backend tests
make test

# Generate mocks (after interface changes)
make mock
```

## Frontend Tests

```bash
cd web
npx vitest run
```
