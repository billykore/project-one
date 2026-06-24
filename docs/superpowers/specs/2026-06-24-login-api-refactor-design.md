# Login Page API Refactor — Design Spec

**Date:** 2026-06-24
**Status:** Approved

## Overview

Refactor the login page to consume the Next.js API route (`/api/login/route.ts`) instead of calling the Go backend directly. Extract the inline Zod schema from the page into the shared `schema.ts`.

## Current State

| File | Issue |
|------|-------|
| `web/app/(auth)/api/login/route.ts` | Skeleton — imports from `schema.ts` but body is empty |
| `web/app/(auth)/api/login/schema.ts` | ✅ Complete — `loginRequestBodySchema` + `LoginRequestBody` type |
| `web/app/(auth)/login/page.tsx` | Inline duplicate of login schema; calls Go backend directly via `api.post("/api/v1/auth/login", ...)` |

## Design

### Architecture

```
Browser (login page)
  │  POST /api/login  (Next.js API route)
  ▼
route.ts (validates body with shared schema, proxies to backend)
  │  POST /api/v1/auth/login  (Go backend)
  ▼
Go Backend → sets HttpOnly cookies → response relayed back
```

### 1. Route (`route.ts`) — Transparent Proxy

- Parse incoming JSON body
- Validate with `loginRequestBodySchema.safeParse()`
- On validation failure: return `400` with first Zod error message
- On success: forward to `{API_URL}/api/v1/auth/login` via `fetch`
- Relay backend response as-is: same status code, same JSON body

### 2. Schema (`schema.ts`) — No Changes

Already contains the canonical schema. No modifications needed.

### 3. Login Page (`page.tsx`) — Two Changes

**Remove:**
- Inline `loginSchema` Zod object and `LoginFormData` type
- `import { z, ZodError } from "zod"`
- `import { api, ApiError } from "@/lib/api"` → keep `ApiError` only

**Add:**
- Import `loginRequestBodySchema` and `LoginRequestBody` from `@/app/(auth)/api/login/schema`
- Alias `LoginRequestBody` as `LoginFormData`
- Replace `api.post()` call with `fetch("/api/login", ...)`
- Parse error response body for error messages (maintains existing error UX)

### Data Flow

```
User submits form
  → validate() uses shared loginRequestBodySchema
  → POST /api/login with { email, password }
  → route.ts validates and proxies to Go backend
  → Backend responds with { message, username } + Set-Cookie headers
  → route.ts relays response
  → Page handles:
      - 200 → router.push("/")
      - 400/401 → show inline error message
      - 500/network error → show error modal
```

### Error Handling

| Scenario | Where Handled | UX |
|----------|--------------|-----|
| Invalid request body (Zod) | `route.ts` → 400 | Inline error on page |
| Bad credentials (backend) | Backend → 401 → relayed | Inline error on page |
| Backend down / 500 | `route.ts` relays → page catch | Error modal |
| Network error (fetch fails) | Page catch block | Error modal |

### Testing Considerations

- **Route**: Test validation (invalid email, short password, missing fields), test successful proxy (mock backend), test error relay (backend 401, 500)
- **Page**: Test form submission calls `/api/login`, test inline error display, test modal error display

## Scope Boundaries

**In scope:**
- Implement `route.ts` transparent proxy
- Update `page.tsx` to use shared schema and call `/api/login`

**Out of scope:**
- Other auth pages (register, logout, change password)
- Backend changes
- E2E tests
