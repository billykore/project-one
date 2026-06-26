# Registration Page Design Spec

**Date:** 2026-06-26
**Status:** Approved — ready for implementation plan

## Goal

Create a registration page (`/register`) for new user sign-up, and update the login page (`/login`) to link to it and show a post-registration success banner.

## Architecture

Single client component following the existing login page pattern. Uses Zod validation, `InputField` components, `ApiError` for error handling, and `useErrorModal` for unexpected errors. The `/api/register` route handler already exists and proxies to `POST /auth/register` on the Go backend.

## API Contract

**Endpoint:** `POST /api/register` → `POST /auth/register`

**Request body (schema):**
```typescript
{
  email: string;        // required, valid email
  first_name: string;   // required, min 3 chars
  last_name: string;    // required, min 3 chars
  username: string;     // required, min 3 chars
  password: string;     // required, min 8 chars
}
```

**Responses:**
- `201` — `{ message: string }` — registration successful
- `400` — `{ error: string }` — validation/duplicate error (e.g., "email already exists")
- `500` — unexpected server error
- Network error — `useErrorModal` path

**Additional client-side field:** `confirmPassword` — must match `password`, not sent to backend.

## Data Flow

```
Registration Form (client component)
  → Zod validation (schema.ts, includes confirmPassword refinement)
    → fetch POST /api/register
      → proxyToBackend → POST /auth/register (Go)
        → 201 → router.push("/login?registered=true")
        → 400 → show inline general error (e.g., duplicate email/username)
        → 500 / network → useErrorModal
```

## Files

| File | Action | Responsibility |
|------|--------|---------------|
| `web/app/api/register/schema.ts` | Create | Zod schema: 5 backend fields + `confirmPassword` with `.refine()` match |
| `web/app/register/page.tsx` | Create | Client component: form with 6 `InputField`s, submit handler, error display |
| `web/app/login/page.tsx` | Modify | Add `?registered=true` success banner + "Create account" link |

## Page Layout

```
┌──────────────────────────────────┐
│     Create your account           │
│                                   │
│  [First Name            ]        │
│  [Last Name             ]        │
│  [Username              ]        │
│  [Email address         ]        │
│  [Password              ]        │
│  [Confirm password      ]        │
│                                   │
│  [  Create account  ]             │
│                                   │
│  Already have an account? Sign in │
└──────────────────────────────────┘
```

- Same card styling as login: `max-w-md`, `space-y-8`, white bg, `rounded-xl shadow-lg`
- Button states: `"Create account"` / `"Creating account..."` (disabled)
- Footer link: `"Already have an account? Sign in"` → `/login`

## Login Page Modifications

1. **Registered success banner** — When `searchParams.registered === "true"`, show a green banner at top of card:
   > ✅ Registration successful! Please sign in with your new account.

2. **Register link** — Below the forgot-password link, add:
   > Don't have an account? Create one → links to `/register`

## Edge Cases

- **Duplicate email/username**: Backend returns 400 — displayed as `general` error (same pattern as login's incorrect credentials)
- **Network failure**: Caught by catch block → `useErrorModal` (unexpected errors)
- **Already authenticated**: No guard on `/register` (follows login page convention — no auth check)
- **Empty form**: Zod handles all required fields with per-field messages
- **confirmPassword mismatch**: Zod `.refine()` catches it before submission — shows field-level error

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Post-registration redirect | `/login?registered=true` | Backend doesn't set auth cookies on register; user must sign in manually |
| Password confirmation | Client-side `confirmPassword` field | Prevents typos; not sent to backend |
| Success feedback | Query param + banner on login | Clear feedback loop; simple to implement |
| Validation | Zod schema (separate file) | Matches existing login pattern |
| Error handling | Inline general + useErrorModal | Matches existing login pattern |
