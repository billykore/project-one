# Navbar Regression Test Report

Date: 2026-06-20
Tester: GitHub Copilot (MCP browser + local servers)
Branch: main

## Objective

Read the last 2 git commits in this branch and test changes for breaking behavior, focused on login and navbar-related flows.

## Commits Reviewed

1. d0a00327bd9674310d51fa79ceb632642c6075c3

- Message: Merge pull request #85 from billykore/refactor/web

1. 9efb693fea50c5414731c0e9e54652063e47c676

- Message: refactor(web): standardize file naming conventions, update import paths, and adjust project directory structure
- Scope (high impact):
  - Navbar/layout: `web/components/layout/navbar.tsx`, `web/components/layout/profile-dropdown.tsx`
  - Notifications: `web/components/notification/*`
  - Posts/pages: `web/app/home/page.tsx`, `web/app/posts/[id]/page.tsx`, `web/app/posts/page.tsx`, `web/app/posts/create/page.tsx`
  - Profile: `web/app/[username]/page.tsx`, `web/components/profile/*`
  - Hooks/tests/types updates under `web/hooks`, `web/tests`, `web/lib/types`

## Environment

- Backend server: `make run` (Go API on :8080)
- Frontend server: `npm run dev` in `web/` (Next.js on :3000)
- Browser automation: MCP browser

## Test Credentials

- Email: <geralt@gmail.com>
- Password: p@ssw0Rd

## Test Steps and Results

1. Open `http://localhost:3000/login`.

- Result: PASS (login page loads)

1. Login with test credentials.

- Result: PASS (redirected to `/home` dashboard)

1. Validate navbar on dashboard.

- Checked: logo/home link, Create Post link, Notifications button, profile button.
- Result: PASS (all visible and interactive)

1. Open notifications dropdown from navbar.

- Result: PASS (panel opens; shows "Offline" and empty state without crash)

1. Open profile dropdown from navbar.

- Checked menu items: Dashboard, View Profile, All Posts, Create New Post, Log Out.
- Result: PASS (menu renders and items present)

1. Navigate via profile dropdown to View Profile (`/geralt`).

- Result: PASS (profile page loads with navbar and profile content)

1. Navigate to Create Post (`/posts/create`).

- Result: PASS (create post form loads)

1. Navigate to Posts list (`/posts`) and a post detail (`/posts/25`).

- Result: PASS (list/detail pages load; detail navbar and interactions render)

## Observations (Non-blocking)

1. Frontend logs repeated `GET /api/ws-token 404` during pages where notifications are initialized.

- Impact observed: notifications panel shows "Offline".
- Breaking status: NOT blocking login/navbar navigation, but should be investigated.

1. Backend logs show multiple slow SQL warnings (200ms+ to ~2s) for user/posts/notifications/likes queries.

- Impact observed: no functional failures during this smoke test.
- Breaking status: NOT a functional break, but a performance concern.

## Verdict

- Login flow: PASS
- Navbar core interactions: PASS
- Route navigation from navbar/profile menu: PASS
- Breaking changes detected in tested paths: NONE

## Recommendation

- Follow-up on `/api/ws-token` 404 behavior to restore online notification state.
- Profile and optimize slow SQL queries to reduce UI latency under load.
