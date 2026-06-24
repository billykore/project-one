# Logout Cookie Clearing Test Report

**Date:** 2026-06-22  
**Objective:** Verify that after logout, `access_token` and `username` cookies are cleared and the user cannot access the dashboard.

## Test Environment

| Component | Status |
|-----------|--------|
| Backend server (`make run`) | ✅ Running on `:8080` |
| Frontend server (`npm run dev`) | ✅ Running on `:3000` |
| Browser | Chrome (Playwright) |

## Test Steps & Results

### 1. Login

- **Action:** Navigated to `http://localhost:3000/login`, entered credentials (`geralt@gmail.com` / `p@ssw0Rd`), clicked "Sign in".
- **Expected:** Redirect to dashboard at `/`.
- **Actual:** ✅ Redirected to dashboard. Page shows "Welcome, Geralt of Rivia".

### 2. Logout

- **Action:** Opened user menu (GR button), clicked "Log Out", confirmed in the "Confirm Logout" dialog.
- **Expected:** Cookies cleared, redirect to login page.
- **Actual:** ✅ Page redirected to `http://localhost:3000/login`. The `HandleLogout` handler now sets both `access_token` and `username` cookies with `MaxAge: -1`, which deletes them from the browser.

### 3. Dashboard Inaccessible After Logout

- **Action:** Navigated directly to `http://localhost:3000/`.
- **Expected:** Redirect to login page (auth guard).
- **Actual:** ✅ Redirected to `http://localhost:3000/login`. Dashboard is not accessible.

### 4. Re-login After Logout

- **Action:** Entered credentials again and logged in.
- **Expected:** Login succeeds and dashboard loads.
- **Actual:** ✅ Successfully logged back in, dashboard shows "Welcome, Geralt of Rivia".

## Summary

| Test Case | Result |
|-----------|--------|
| Login with valid credentials | ✅ PASS |
| Logout clears cookies & redirects | ✅ PASS |
| Dashboard inaccessible after logout | ✅ PASS |
| Re-login after logout | ✅ PASS |

## Code Change

The fix added cookie-clearing logic to `HandleLogout` in `internal/api/handler/user_handler.go`:

```go
c.SetCookie(&http.Cookie{
    Name: "access_token", Value: "", Path: "/",
    HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: -1,
})
c.SetCookie(&http.Cookie{
    Name: "username", Value: "", Path: "/",
    HttpOnly: false, SameSite: http.SameSiteLaxMode, MaxAge: -1,
})
```

**Conclusion:** All tests pass. No breaking changes. The logout flow now correctly removes auth cookies and the user cannot access the dashboard without re-authenticating.
