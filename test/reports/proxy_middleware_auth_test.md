# Test Report: Proxy Middleware & Auth Changes

**Date:** 2026-06-22  
**Test Objective:** Verify the proxy middleware, login flow, and auth changes work correctly.

---

## Changes Tested

1. **`web/proxy.ts`** — Next.js proxy middleware (replaces deprecated `middleware.ts`) guards all non-public routes, bypasses `/api`, `/_next`, `/static` paths.
2. **`web/hooks/use-user.ts`** — New reusable hook for fetching authenticated user data.
3. **`web/app/page.tsx`** — Simplified to use `useUser()` hook.
4. **`internal/api/handler/user_handler.go`** — Fixed cookie `MaxAge` (was `ExpiresAt.Second()` → now `int(time.Until(...).Seconds())`), made `username` cookie non-HttpOnly, added username to response body.
5. **`internal/core/usecase/login_usecase.go`** — Fixed missing `accessToken.Username` before `StoreToken`.
6. **`web/hooks/use-profile-controller.ts`** — Replaced `localStorage` with cookie read.

---

## Test Results

### 1. Proxy Middleware — Unauthenticated Access

| Test | Result |
|------|--------|
| Access `/` without cookies | ✅ Redirects to `/login` |
| Access `/login` (public route) | ✅ Renders login page |
| API requests to `/api/v1/...` bypass proxy | ✅ Pass through to backend |

### 2. Login Flow (Form Submission)

| Step | Result |
|------|--------|
| Fill email (`geralt@gmail.com`) and password | ✅ |
| Click "Sign in" | ✅ API call to `/api/v1/auth/login` returns 200 |
| Redirect to dashboard | ✅ Navigates to `/` |

### 3. Dashboard Display

| Element | Result |
|---------|--------|
| Welcome message | ✅ "Welcome, Geralt of Rivia" |
| Email display | ✅ `geralt@gmail.com` |
| Dashboard links | ✅ View Your Profile, View All Your Posts, Create a New Post |
| Navbar | ✅ Shows with Dashboard title and Log In link |

### 4. Cookie Behavior

| Cookie | HttpOnly | Max-Age | Result |
|--------|----------|---------|--------|
| `access_token` | ✅ Yes | ~900s (15m) | ✅ Set correctly |
| `username` | ❌ No (readable by JS) | ~900s (15m) | ✅ Set and readable via `document.cookie` |

### 5. API Response

```json
{"message":"Login successful","username":"geralt"}
```

---

## Bugs Fixed During Testing

| # | Bug | Fix |
|---|-----|-----|
| 1 | Cookie `MaxAge` set to second-of-minute (0-59) instead of duration | `int(time.Until(accessToken.ExpiresAt).Seconds())` |
| 2 | `username` cookie was HttpOnly, blocking client-side redirect guard | Changed to `HttpOnly: false` |
| 3 | `username` field missing from login response body | Added `Username: accessToken.Username` to response |
| 4 | `lib/proxy.ts` was never wired as middleware (Next.js only reads `middleware.ts` at root) | Created `web/proxy.ts` with proper `export function proxy` |
| 5 | Proxy wasn't bypassing `/api` routes, blocking login requests | Added `pathname.startsWith("/api")` to bypass condition |
| 6 | `accessToken.Username` was empty before `StoreToken` in login usecase | Added `accessToken.Username = user.Username` |

---

## Conclusion

All changes work as expected. The proxy middleware correctly:

- Redirects unauthenticated users to `/login`
- Allows API calls to pass through
- Protects the dashboard behind authentication

The login flow completes successfully, cookies are set with proper expiration, and the dashboard displays user data correctly.

---

**Screenshots:** See `test/screenshots/`
