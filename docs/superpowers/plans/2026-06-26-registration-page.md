# Registration Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a `/register` page for new user sign-up and update `/login` with a register link and post-registration success banner.

**Architecture:** Follows the existing login page pattern — client component with Zod validation, `InputField`, `ApiError`, and `useErrorModal`. The `/api/register` proxy route already exists.

**Tech Stack:** Next.js 16 (App Router), React 19, TypeScript, Zod, Tailwind CSS 4

## Global Constraints

- Follow existing patterns from `web/app/login/page.tsx` for component structure, error handling, and styling
- Zod validation schemas live in `web/app/api/<endpoint>/schema.ts`
- API errors use `ApiError` from `web/lib/errors.ts`
- Unexpected errors use `useErrorModal` hook
- Form inputs use `InputField` from `web/components/ui/input.tsx`
- Client components use `"use client"` directive

---

## File Structure

| File | Action | Responsibility |
|------|--------|---------------|
| `web/app/api/register/schema.ts` | Create | Zod schema with 5 backend fields + `confirmPassword` refinement |
| `web/app/register/page.tsx` | Create | Client component: registration form |
| `web/app/login/page.tsx` | Modify | Add `registered` success banner + register link |

---

### Task 1: Registration Zod Schema

**Files:**
- Create: `web/app/api/register/schema.ts`

**Interfaces:**
- Consumes: none
- Produces: `registerRequestBodySchema` (ZodSchema), `RegisterRequestBody` (type)

- [ ] **Step 1: Create the schema file**
  - Define `registerRequestBodySchema` using `z.object({...})`
  - Fields: `email` (string, required, valid email), `first_name` (string, min 3), `last_name` (string, min 3), `username` (string, min 3), `password` (string, min 8), `confirmPassword` (string)
  - Add `.refine()` on the object: password === confirmPassword, with message "Passwords do not match"
  - Export `RegisterRequestBody` type via `z.infer<typeof registerRequestBodySchema>`
  - Follow the exact style of `web/app/api/login/schema.ts`

- [ ] **Step 2: Verify schema exports**
  - Confirm the file exports both the schema and the inferred type
  - Confirm `confirmPassword` is omitted from the type sent to the API (handled in the submit handler)

---

### Task 2: Registration Page Component

**Files:**
- Create: `web/app/register/page.tsx`

**Interfaces:**
- Consumes: `registerRequestBodySchema`, `RegisterRequestBody` from `web/app/api/register/schema`
- Consumes: `InputField` from `web/components/ui/input`
- Consumes: `ApiError` from `web/lib/errors`
- Consumes: `useErrorModal` from `web/hooks/use-error-modal`

- [ ] **Step 1: Scaffold the page component**
  - Create `"use client"` component `RegisterPage`
  - Set up state: `formData` (RegisterRequestBody with empty strings), `errors` (Record of field → string), `isSubmitting` (boolean)
  - Import router from `next/navigation`, Link from `next/link`

- [ ] **Step 2: Implement handleChange**
  - Update `formData` on input change
  - Clear field error when user starts typing in that field
  - Follow login page's `handleChange` pattern exactly

- [ ] **Step 3: Implement validate function**
  - Parse `formData` with `registerRequestBodySchema`
  - Map Zod issues to per-field errors (same pattern as login page)
  - Return boolean

- [ ] **Step 4: Implement handleSubmit**
  - Prevent default, call validate()
  - POST to `/api/register` with: email, first_name, last_name, username, password (omit confirmPassword)
  - On 201: `router.push("/login?registered=true")`
  - On 400: set `general` error with `err.message`
  - On other errors: call `showError` from `useErrorModal`
  - Finally: set `isSubmitting` to false

- [ ] **Step 5: Render the form UI**
  - Card layout matching login page (`min-h-screen flex items-center justify-center`, `max-w-md`, `bg-white rounded-xl shadow-lg`)
  - Title: "Create your account"
  - 6 InputField components: First Name, Last Name, Username, Email, Password (type=password, autoComplete=new-password), Confirm Password (type=password, autoComplete=new-password)
  - General error banner (red, same as login) when `errors.general` is set
  - Submit button: "Create account" / "Creating account..." with disabled state
  - Footer: "Already have an account? Sign in" → `<Link href="/login">`

---

### Task 3: Update Login Page

**Files:**
- Modify: `web/app/login/page.tsx`

**Interfaces:**
- Consumes: Next.js `searchParams` prop for the `registered` query param
- Produces: Updated login page with success banner + register link

- [ ] **Step 1: Add searchParams prop and success banner**
  - Add `searchParams` prop to `LoginPage` (type: `{ registered?: string }`)
  - At the top of the card (before the title), conditionally render a green success banner when `searchParams.registered === "true"`
  - Banner text: "✅ Registration successful! Please sign in with your new account."
  - Use green styling: `bg-green-50 border border-green-200 text-green-800 rounded-md p-4`

- [ ] **Step 2: Add register link**
  - Below the forgot-password link, add:
  - Text: "Don't have an account? Create one" with a `<Link href="/register">` on "Create one"
  - Style matching the existing forgot-password link pattern

---

### Task 4: Verification

- [ ] **Step 1: Verify the registration page renders**
  - Run `npm run dev` in `web/`
  - Navigate to `http://localhost:3000/register`
  - Confirm the form renders with all 6 fields

- [ ] **Step 2: Verify client-side validation**
  - Submit empty form → all required field errors appear
  - Enter mismatched passwords → "Passwords do not match" error
  - Enter email without @ → email validation error

- [ ] **Step 3: Verify login page changes**
  - Confirm "Don't have an account? Create one" link appears on `/login`
  - Confirm link navigates to `/register`
  - Navigate to `/login?registered=true` → confirm green success banner

- [ ] **Step 4: Verify end-to-end flow (with backend running)**
  - Register a new user → redirected to `/login?registered=true`
  - See success banner on login page
  - Sign in with new credentials → successfully logged in

---

### Task 5: Commit

- [ ] **Step 1: Commit all changes**
  - `git add web/app/api/register/schema.ts web/app/register/page.tsx web/app/login/page.tsx docs/superpowers/specs/2026-06-26-registration-page-design.md docs/superpowers/plans/`
  - Commit message: `feat: add registration page with validation and login page updates`
