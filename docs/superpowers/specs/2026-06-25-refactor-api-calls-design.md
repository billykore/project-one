# Refactor Frontend API Calls to Next.js API Routes

**Date:** 2026-06-25
**Status:** Approved

## Goal

Eliminate the `api` / `apiServer` abstraction layer (`lib/api.ts`, `lib/api-server.ts`) and the catch-all proxy (`app/api/proxy/`). Every frontend call — server components, client components, hooks, server actions — uses raw `fetch("/api/*")` to a dedicated Next.js API route. Each route is a thin proxy that forwards the request and user cookies to the Go backend.

## Motivation

- Single, consistent API call pattern across the entire frontend
- No runtime proxy routing or path-mapping logic
- Each API route is a tiny, focused file — easy to read and modify
- Removes `NEXT_PUBLIC_API_URL` / `API_URL` environment variable complexity

## Architecture

```
┌──────────────────────┐     fetch("/api/*")     ┌─────────────────────┐     HTTP + Cookie     ┌──────────────────┐
│  Server Components   │ ──────────────────────→ │                     │ ───────────────────→  │                  │
│  Client Components   │ ──────────────────────→ │  Next.js API Routes │                      │   Go Backend     │
│  Hooks               │ ──────────────────────→ │  (thin proxy layer) │                      │   :8080          │
│  Server Actions      │ ──────────────────────→ │                     │                      │                  │
└──────────────────────┘                         └─────────────────────┘                      └──────────────────┘
```

## What Gets Removed

- `web/lib/api.ts` — the `api` object, `ApiError` class, `BASE_URL` logic
- `web/lib/api-server.ts` — the `apiServer` object
- `web/app/api/proxy/[...path]/route.ts` — the catch-all proxy

## What Gets Created

### Shared Helper: `web/lib/api-proxy.ts`

Used by every API route handler. Reads cookies from the incoming `Request`, forwards them to the Go backend as a `Cookie` header, and returns the backend response with any `Set-Cookie` headers preserved.

### Shared Utility: `web/lib/errors.ts`

Contains the `ApiError` class (moved from `lib/api.ts`). Also contains a `handleApiResponse<T>(response: Response): Promise<T>` helper for JSON parsing and error extraction — shared between client call sites.

### Server Cookie Forwarding: `web/lib/server-fetch.ts`

For server components / server actions, reads `cookies()` from `next/headers` and passes them to the API route `fetch` call so auth cookies reach the Next.js API route layer.

### API Routes (16 files, RESTful structure)

```
web/app/api/
├── login/route.ts                         ← POST    → backend /auth/login
├── logout/route.ts                        ← POST    → backend /auth/logout
├── register/route.ts                      ← POST    → backend /auth/register
├── posts/
│   ├── route.ts                           ← GET,POST → backend /posts
│   └── [id]/
│       ├── route.ts                       ← GET,PUT,DELETE → /posts/:id
│       ├── likes/route.ts                 ← GET,POST,DELETE → /posts/:id/likes
│       └── comments/route.ts              ← POST → /posts/:id/comments
├── comments/
│   └── [id]/route.ts                      ← PUT,DELETE → /comments/:id
├── users/
│   ├── [username]/
│   │   ├── route.ts                       ← GET → /users/:username
│   │   ├── followers/route.ts             ← GET,POST,DELETE → /users/:username/followers
│   │   ├── following/route.ts             ← GET → /users/:username/following
│   │   └── posts/route.ts                 ← GET → /users/:username/posts
│   └── password/route.ts                  ← PUT → /users/password
└── notifications/
    ├── route.ts                           ← GET → /notifications
    ├── [id]/
    │   └── read/route.ts                  ← PUT → /notifications/:id/read
    └── read-all/route.ts                  ← PUT → /notifications/read-all
```

## What Gets Modified — Call Sites

### Client Components & Hooks

Replace `import { api } from "@/lib/api"` with `import { handleApiResponse } from "@/lib/errors"`.

| File | Current | New |
|------|---------|-----|
| `components/layout/profile-dropdown.tsx` | `api.post("/api/v1/auth/logout", {})` | `fetch("/api/logout", { method: "POST" })` |
| `components/posts/post-interaction-section.tsx` | Multiple `api.get/post/put/delete` | `fetch("/api/posts/...", { ... })` |
| `components/posts/delete-post-button.tsx` | `api.delete(...)` | `fetch(..., { method: "DELETE" })` |
| `hooks/use-profile-controller.ts` | `api.delete/post/put` | `fetch("/api/users/...", { ... })` |
| `lib/notifications-api.ts` | `api.get/put` | `fetch("/api/notifications/...", { ... })` |

### Server Components & Actions

Replace `import { apiServer } from "@/lib/api-server"` with `import { serverFetch } from "@/lib/server-fetch"`.

| File | Current | New |
|------|---------|-----|
| `app/posts/page.tsx` | `apiServer.get<Post[]>("/posts")` | `serverFetch("/api/posts").then(r => r.json())` |
| `app/posts/[id]/page.tsx` | `apiServer.get<Post>(...)` | `serverFetch("/api/posts/${id}").then(...)` |
| `app/[username]/page.tsx` | 4 × `apiServer.get` | 4 × `serverFetch("/api/users/...")` |
| `app/posts/create/actions.ts` | `apiServer.post(...)` | `serverFetch("/api/posts", { method: "POST", ... })` |

## Auth Handling

- API routes read `access_token` cookie from the incoming `Request` and forward it as a `Cookie` header to the Go backend
- Go backend remains the sole authority for authentication — returns 401 if token is invalid/expired
- API routes that receive a `Set-Cookie` header from the backend forward it to the client
- No JWT validation at the Next.js layer

## Error Handling

- `handleApiResponse<T>(response)` checks `response.ok`, parses JSON error bodies, and throws `ApiError` with the backend's error message and status code
- Client call sites wrap `fetch` + `handleApiResponse` in try/catch as before
- `ApiError` is the single error type for all API calls
