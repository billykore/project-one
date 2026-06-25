# Refactor Frontend API Calls to Next.js API Routes — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace `api`/`apiServer` abstractions with raw `fetch("/api/*")` calls through dedicated Next.js API route proxies that forward cookies to the Go backend.

**Architecture:** All frontend code calls `fetch("/api/*")` → Next.js API route reads cookies from incoming request → forwards as `Cookie` header to Go backend at `localhost:8080` → returns backend response with any `Set-Cookie` headers. Three shared helpers: `api-proxy.ts` (used by route handlers), `errors.ts` (ApiError + response parsing), `server-fetch.ts` (server component fetch with cookie forwarding).

**Tech Stack:** Next.js 16 App Router, TypeScript, Zod (existing schemas), Go backend at `localhost:8080`

## Global Constraints

- Go backend URL: `process.env.API_URL || "http://localhost:8080"` (used only in API route handlers)
- All API routes forward `access_token` cookie from incoming request to backend as `Cookie` header
- All API routes forward `Set-Cookie` from backend response to client
- No JWT validation at the Next.js layer — backend is sole auth authority
- `ApiError` is the single error type for all API calls
- Keep existing Zod schemas in `app/api/login/schema.ts` and `app/api/posts/schema.ts`

---

## File Map

```
Create:
  web/lib/errors.ts                    — ApiError class + handleApiResponse helper
  web/lib/api-proxy.ts                 — proxyToBackend helper for route handlers
  web/lib/server-fetch.ts              — serverFetch for server components/actions
  web/app/api/logout/route.ts          — POST /api/logout
  web/app/api/register/route.ts        — POST /api/register
  web/app/api/posts/[id]/route.ts      — GET,PUT,DELETE /api/posts/:id
  web/app/api/posts/[id]/likes/route.ts — GET,POST,DELETE /api/posts/:id/likes
  web/app/api/posts/[id]/comments/route.ts — POST /api/posts/:id/comments
  web/app/api/comments/[id]/route.ts   — PUT,DELETE /api/comments/:id
  web/app/api/users/[username]/route.ts — GET /api/users/:username
  web/app/api/users/[username]/followers/route.ts — GET,POST,DELETE
  web/app/api/users/[username]/following/route.ts — GET
  web/app/api/users/[username]/posts/route.ts — GET
  web/app/api/users/password/route.ts  — PUT /api/users/password
  web/app/api/notifications/route.ts   — GET /api/notifications
  web/app/api/notifications/[id]/read/route.ts — PUT
  web/app/api/notifications/read-all/route.ts  — PUT

Modify:
  web/app/api/login/route.ts           — use proxyToBackend helper
  web/app/api/posts/route.ts           — use proxyToBackend helper
  web/app/login/page.tsx               — already uses fetch("/api/login") ✅ no change
  web/app/posts/page.tsx               — replace apiServer with serverFetch
  web/app/posts/[id]/page.tsx          — replace apiServer with serverFetch
  web/app/[username]/page.tsx          — replace apiServer with serverFetch
  web/app/posts/create/actions.ts      — replace apiServer with serverFetch
  web/components/layout/profile-dropdown.tsx — replace api.* with fetch + handleApiResponse
  web/components/posts/post-interaction-section.tsx — replace api.* with fetch + handleApiResponse
  web/components/posts/delete-post-button.tsx — replace api.* with fetch + handleApiResponse
  web/hooks/use-profile-controller.ts  — replace api.* with fetch + handleApiResponse
  web/lib/notifications-api.ts         — replace api.* with fetch + handleApiResponse

Delete:
  web/lib/api.ts
  web/lib/api-server.ts
  web/app/api/proxy/[...path]/route.ts
  web/app/api/proxy/[...path]/         (directory)
  web/app/api/posts/schema.ts          — no longer imported (route.ts replaced)
  web/app/api/posts/model.ts           — unused
```

---

### Task 1: Create Shared Helper — `lib/errors.ts`

**Files:**
- Create: `web/lib/errors.ts`
- The file must export `ApiError` class and `handleApiResponse<T>` function.

**Interfaces:**
- Produces: `export class ApiError extends Error { status: number; constructor(message: string, status: number) }`
- Produces: `export async function handleApiResponse<T>(response: Response): Promise<T>`

- [ ] **Step 1: Write the file**

```ts
export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
    this.name = "ApiError";
  }
}

export async function handleApiResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    let errorMessage = `Something went wrong! (${response.status})`;
    const contentType = response.headers.get("content-type");

    if (contentType && contentType.includes("application/json")) {
      const errorData = await response.json().catch(() => ({}));
      if (errorData.error) errorMessage = errorData.error;
    }

    throw new ApiError(errorMessage, response.status);
  }

  const contentType = response.headers.get("content-type");
  if (contentType && contentType.includes("application/json")) {
    return response.json();
  }

  return {} as T;
}
```

- [ ] **Step 2: Verify the file has no TypeScript errors**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npx tsc --noEmit --pretty lib/errors.ts 2>&1 | head -20`
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add web/lib/errors.ts
git commit -m "feat: add shared ApiError and handleApiResponse to lib/errors.ts"
```

---

### Task 2: Create Shared Helper — `lib/api-proxy.ts`

**Files:**
- Create: `web/lib/api-proxy.ts`

**Interfaces:**
- Produces: `export async function proxyToBackend(req: Request, backendPath: string): Promise<Response>`

- [ ] **Step 1: Write the file**

```ts
const BACKEND_URL = process.env.API_URL || "http://localhost:8080";

export async function proxyToBackend(req: Request, backendPath: string): Promise<Response> {
  let body: string | undefined;
  const contentType = req.headers.get("content-type") || "";

  if (contentType.includes("application/json")) {
    try {
      body = await req.text();
    } catch {
      // no body
    }
  }

  const cookieHeader = req.headers.get("cookie") || "";

  let backend: Response;
  try {
    backend = await fetch(`${BACKEND_URL}${backendPath}`, {
      method: req.method,
      headers: {
        "Content-Type": contentType || "application/json",
        ...(cookieHeader ? { Cookie: cookieHeader } : {}),
      },
      ...(body ? { body } : {}),
    });
  } catch {
    return Response.json({ error: "Unable to connect to the server" }, { status: 502 });
  }

  let data: unknown;
  try {
    data = await backend.json();
  } catch {
    data = { error: "Invalid response from server" };
  }

  const responseHeaders: Record<string, string> = {};
  const setCookie = backend.headers.get("set-cookie");
  if (setCookie) {
    responseHeaders["Set-Cookie"] = setCookie;
  }

  return Response.json(data, { status: backend.status, headers: responseHeaders });
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npx tsc --noEmit --pretty lib/api-proxy.ts 2>&1 | head -20`
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add web/lib/api-proxy.ts
git commit -m "feat: add proxyToBackend helper for API route handlers"
```

---

### Task 3: Create Shared Helper — `lib/server-fetch.ts`

**Files:**
- Create: `web/lib/server-fetch.ts`

**Interfaces:**
- Produces: `export async function serverFetch(path: string, init?: RequestInit): Promise<Response>`

- [ ] **Step 1: Write the file**

```ts
import { cookies, headers } from "next/headers";

export async function serverFetch(path: string, init?: RequestInit): Promise<Response> {
  const [cookieStore, headersList] = await Promise.all([cookies(), headers()]);
  const cookieString = cookieStore.toString();
  const host = headersList.get("host") || "localhost:3000";
  const protocol = process.env.NODE_ENV === "production" ? "https" : "http";
  const baseUrl = `${protocol}://${host}`;

  return fetch(`${baseUrl}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
      ...(cookieString ? { Cookie: cookieString } : {}),
    },
  });
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npx tsc --noEmit --pretty lib/server-fetch.ts 2>&1 | head -20`
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add web/lib/server-fetch.ts
git commit -m "feat: add serverFetch helper for server components"
```

---

### Task 4: Create & Update Auth API Routes

**Files:**
- Modify: `web/app/api/login/route.ts` — rewrite to use `proxyToBackend`
- Create: `web/app/api/logout/route.ts`
- Create: `web/app/api/register/route.ts`

**Interfaces:**
- Consumes: `proxyToBackend` from `@/lib/api-proxy`
- Produces: `POST /api/login`, `POST /api/logout`, `POST /api/register` route handlers

- [ ] **Step 1: Rewrite `web/app/api/login/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function POST(req: Request) {
  return proxyToBackend(req, "/auth/login");
}
```

- [ ] **Step 2: Create `web/app/api/logout/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function POST(req: Request) {
  return proxyToBackend(req, "/auth/logout");
}
```

- [ ] **Step 3: Create `web/app/api/register/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function POST(req: Request) {
  return proxyToBackend(req, "/auth/register");
}
```

- [ ] **Step 4: Verify TypeScript compiles for all three**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npx tsc --noEmit --pretty app/api/login/route.ts app/api/logout/route.ts app/api/register/route.ts 2>&1 | head -20`
Expected: No errors.

- [ ] **Step 5: Commit**

```bash
git add web/app/api/login/route.ts web/app/api/logout/route.ts web/app/api/register/route.ts
git commit -m "feat: add auth API routes (login, logout, register) using proxyToBackend"
```

---

### Task 5: Create Posts API Routes

**Files:**
- Modify: `web/app/api/posts/route.ts` — rewrite to use `proxyToBackend`
- Create: `web/app/api/posts/[id]/route.ts`
- Create: `web/app/api/posts/[id]/likes/route.ts`
- Create: `web/app/api/posts/[id]/comments/route.ts`

**Interfaces:**
- Consumes: `proxyToBackend` from `@/lib/api-proxy`
- Produces: `GET,POST /api/posts`, `GET,PUT,DELETE /api/posts/:id`, `GET,POST,DELETE /api/posts/:id/likes`, `POST /api/posts/:id/comments`

- [ ] **Step 1: Rewrite `web/app/api/posts/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(req: Request) {
  return proxyToBackend(req, "/posts");
}

export async function POST(req: Request) {
  return proxyToBackend(req, "/posts");
}
```

- [ ] **Step 2: Create `web/app/api/posts/[id]/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(_req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return proxyToBackend(_req, `/posts/${id}`);
}

export async function PUT(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return proxyToBackend(req, `/posts/${id}`);
}

export async function DELETE(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return proxyToBackend(req, `/posts/${id}`);
}
```

- [ ] **Step 3: Create `web/app/api/posts/[id]/likes/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(_req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return proxyToBackend(_req, `/posts/${id}/likes`);
}

export async function POST(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return proxyToBackend(req, `/posts/${id}/likes`);
}

export async function DELETE(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return proxyToBackend(req, `/posts/${id}/likes`);
}
```

- [ ] **Step 4: Create `web/app/api/posts/[id]/comments/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function POST(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return proxyToBackend(req, `/posts/${id}/comments`);
}
```

- [ ] **Step 5: Verify TypeScript compiles**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npx tsc --noEmit --pretty app/api/posts/route.ts app/api/posts/\[id\]/route.ts app/api/posts/\[id\]/likes/route.ts app/api/posts/\[id\]/comments/route.ts 2>&1 | head -20`
Expected: No errors.

- [ ] **Step 6: Commit**

```bash
git add web/app/api/posts/
git commit -m "feat: add posts API routes (CRUD, likes, comments) using proxyToBackend"
```

---

### Task 6: Create Comments API Route

**Files:**
- Create: `web/app/api/comments/[id]/route.ts`

**Interfaces:**
- Consumes: `proxyToBackend` from `@/lib/api-proxy`
- Produces: `PUT,DELETE /api/comments/:id`

- [ ] **Step 1: Write the file**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function PUT(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return proxyToBackend(req, `/comments/${id}`);
}

export async function DELETE(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return proxyToBackend(req, `/comments/${id}`);
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npx tsc --noEmit --pretty app/api/comments/\[id\]/route.ts 2>&1 | head -20`
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add web/app/api/comments/
git commit -m "feat: add comments API routes (edit, delete) using proxyToBackend"
```

---

### Task 7: Create Users API Routes

**Files:**
- Create: `web/app/api/users/[username]/route.ts`
- Create: `web/app/api/users/[username]/followers/route.ts`
- Create: `web/app/api/users/[username]/following/route.ts`
- Create: `web/app/api/users/[username]/posts/route.ts`
- Create: `web/app/api/users/password/route.ts`

**Interfaces:**
- Consumes: `proxyToBackend` from `@/lib/api-proxy`

- [ ] **Step 1: Create `web/app/api/users/[username]/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(_req: Request, { params }: { params: Promise<{ username: string }> }) {
  const { username } = await params;
  return proxyToBackend(_req, `/users/${username}`);
}
```

- [ ] **Step 2: Create `web/app/api/users/[username]/followers/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(_req: Request, { params }: { params: Promise<{ username: string }> }) {
  const { username } = await params;
  return proxyToBackend(_req, `/users/${username}/followers`);
}

export async function POST(req: Request, { params }: { params: Promise<{ username: string }> }) {
  const { username } = await params;
  return proxyToBackend(req, `/users/${username}/followers`);
}

export async function DELETE(req: Request, { params }: { params: Promise<{ username: string }> }) {
  const { username } = await params;
  return proxyToBackend(req, `/users/${username}/followers`);
}
```

- [ ] **Step 3: Create `web/app/api/users/[username]/following/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(_req: Request, { params }: { params: Promise<{ username: string }> }) {
  const { username } = await params;
  return proxyToBackend(_req, `/users/${username}/following`);
}
```

- [ ] **Step 4: Create `web/app/api/users/[username]/posts/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(_req: Request, { params }: { params: Promise<{ username: string }> }) {
  const { username } = await params;
  return proxyToBackend(_req, `/users/${username}/posts`);
}
```

- [ ] **Step 5: Create `web/app/api/users/password/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function PUT(req: Request) {
  return proxyToBackend(req, "/users/password");
}
```

- [ ] **Step 6: Verify TypeScript compiles**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npx tsc --noEmit --pretty app/api/users/\[username\]/route.ts app/api/users/\[username\]/followers/route.ts app/api/users/\[username\]/following/route.ts app/api/users/\[username\]/posts/route.ts app/api/users/password/route.ts 2>&1 | head -20`
Expected: No errors.

- [ ] **Step 7: Commit**

```bash
git add web/app/api/users/
git commit -m "feat: add users API routes (profile, followers, following, posts, password)"
```

---

### Task 8: Create Notifications API Routes

**Files:**
- Create: `web/app/api/notifications/route.ts`
- Create: `web/app/api/notifications/[id]/read/route.ts`
- Create: `web/app/api/notifications/read-all/route.ts`

**Interfaces:**
- Consumes: `proxyToBackend` from `@/lib/api-proxy`

- [ ] **Step 1: Create `web/app/api/notifications/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(req: Request) {
  return proxyToBackend(req, "/notifications");
}
```

- [ ] **Step 2: Create `web/app/api/notifications/[id]/read/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function PUT(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return proxyToBackend(req, `/notifications/${id}/read`);
}
```

- [ ] **Step 3: Create `web/app/api/notifications/read-all/route.ts`**

```ts
import { proxyToBackend } from "@/lib/api-proxy";

export async function PUT(req: Request) {
  return proxyToBackend(req, "/notifications/read-all");
}
```

- [ ] **Step 4: Verify TypeScript compiles**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npx tsc --noEmit --pretty app/api/notifications/route.ts app/api/notifications/\[id\]/read/route.ts app/api/notifications/read-all/route.ts 2>&1 | head -20`
Expected: No errors.

- [ ] **Step 5: Commit**

```bash
git add web/app/api/notifications/
git commit -m "feat: add notifications API routes (list, mark-read, mark-all-read)"
```

---

### Task 9: Update Client-Side Call Sites

**Files:**
- Modify: `web/components/layout/profile-dropdown.tsx`
- Modify: `web/components/posts/post-interaction-section.tsx`
- Modify: `web/components/posts/delete-post-button.tsx`
- Modify: `web/hooks/use-profile-controller.ts`
- Modify: `web/lib/notifications-api.ts`

**Interfaces:**
- Consumes: `handleApiResponse`, `ApiError` from `@/lib/errors`
- Removes dependency on `@/lib/api`

- [ ] **Step 1: Update `web/components/layout/profile-dropdown.tsx`**

Replace `import { api } from "@/lib/api";` with `import { handleApiResponse } from "@/lib/errors";` and replace the logout call.

Old code (line 5 and lines 65-68):
```ts
import { api } from "@/lib/api";
```
```ts
    try {
      await api.post("/api/v1/auth/logout", {});
```

New code:
```ts
import { handleApiResponse } from "@/lib/errors";
```
```ts
    try {
      await handleApiResponse(await fetch("/api/logout", { method: "POST" }));
```

- [ ] **Step 2: Update `web/components/posts/post-interaction-section.tsx`**

Replace `import { api } from "@/lib/api";` (line 7) with `import { handleApiResponse } from "@/lib/errors";`.

Replace the like-status fetch (lines 49-51):
```ts
// Old:
        const res = await api.get<{ liked: boolean; like_count: number }>(
          `/api/v1/posts/${postId}/likes`
        );
// New:
        const res = await handleApiResponse<{ liked: boolean; like_count: number }>(
          await fetch(`/api/posts/${postId}/likes`)
        );
```

Replace like toggle (lines 70-73):
```ts
// Old:
      if (originalIsLiked) {
        await api.delete(`/api/v1/posts/${postId}/likes`);
      } else {
        await api.post(`/api/v1/posts/${postId}/likes`, {});
// New:
      if (originalIsLiked) {
        await handleApiResponse(await fetch(`/api/posts/${postId}/likes`, { method: "DELETE" }));
      } else {
        await handleApiResponse(await fetch(`/api/posts/${postId}/likes`, { method: "POST" }));
```

Replace fetchComments (line 83):
```ts
// Old:
      const res = await api.get<{ comments?: Comment[] }>(`/api/v1/posts/${postId}`);
// New:
      const res = await handleApiResponse<{ comments?: Comment[] }>(await fetch(`/api/posts/${postId}`));
```

Replace handleAddComment (lines 96-99):
```ts
// Old:
    await api.post(`/api/v1/posts/${postId}/comments`, {
      content,
      id: postId,
    });
// New:
    await handleApiResponse(await fetch(`/api/posts/${postId}/comments`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ content, id: postId }),
    }));
```

Replace handleEditComment (lines 103-104):
```ts
// Old:
    await api.put(`/api/v1/comments/${commentId}`, { content });
// New:
    await handleApiResponse(await fetch(`/api/comments/${commentId}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ content }),
    }));
```

Replace handleDeleteComment (lines 109-110):
```ts
// Old:
    await api.delete(`/api/v1/comments/${commentId}`);
// New:
    await handleApiResponse(await fetch(`/api/comments/${commentId}`, { method: "DELETE" }));
```

- [ ] **Step 3: Update `web/components/posts/delete-post-button.tsx`**

Replace `import { api, ApiError } from "@/lib/api";` (line 7) with `import { ApiError, handleApiResponse } from "@/lib/errors";`.

Replace the delete call (line 71):
```ts
// Old:
      await api.delete(`/api/v1/posts/${postId}`);
// New:
      await handleApiResponse(await fetch(`/api/posts/${postId}`, { method: "DELETE" }));
```

- [ ] **Step 4: Update `web/hooks/use-profile-controller.ts`**

Replace `import { api, ApiError } from "@/lib/api";` (line 3) with `import { ApiError, handleApiResponse } from "@/lib/errors";`.

Replace follow toggle (lines 40-43):
```ts
// Old:
        await api.delete(`/api/v1/users/${profileUsername}/followers`);
        ...
        await api.post(`/api/v1/users/${profileUsername}/followers`, {});
// New:
        await handleApiResponse(await fetch(`/api/users/${profileUsername}/followers`, { method: "DELETE" }));
        ...
        await handleApiResponse(await fetch(`/api/users/${profileUsername}/followers`, { method: "POST" }));
```

Replace password change (lines 88-91):
```ts
// Old:
      await api.put("/api/v1/users/password", {
        old_password: pwdForm.oldPassword,
        new_password: pwdForm.newPassword,
      });
// New:
      await handleApiResponse(await fetch("/api/users/password", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          old_password: pwdForm.oldPassword,
          new_password: pwdForm.newPassword,
        }),
      }));
```

- [ ] **Step 5: Update `web/lib/notifications-api.ts`**

Replace `import { api } from "./api";` (line 1) with `import { handleApiResponse } from "./errors";`.

Replace fetchNotifications (line 62):
```ts
// Old:
  const payload = await api.get<NotificationsListResponse | BackendNotification[]>(NOTIFICATIONS_ENDPOINT);
// New:
  const payload = await handleApiResponse<NotificationsListResponse | BackendNotification[]>(
    await fetch(`/api${NOTIFICATIONS_ENDPOINT}`)
  );
```

Replace markNotificationAsRead (line 73):
```ts
// Old:
  await api.put(`${NOTIFICATIONS_ENDPOINT}/${id}/read`, {});
// New:
  await handleApiResponse(await fetch(`/api/notifications/${id}/read`, { method: "PUT" }));
```

Replace markAllNotificationsAsRead (line 77):
```ts
// Old:
  await api.put(`${NOTIFICATIONS_ENDPOINT}/read-all`, {});
// New:
  await handleApiResponse(await fetch("/api/notifications/read-all", { method: "PUT" }));
```

Also remove the unused `NOTIFICATIONS_ENDPOINT` constant.

- [ ] **Step 6: Verify TypeScript compiles for all modified files**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npx tsc --noEmit --pretty 2>&1 | head -30`
Expected: No errors related to these files (there may be pre-existing unrelated warnings).

- [ ] **Step 7: Run existing tests to verify nothing is broken**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npm test -- --run 2>&1 | tail -20`
Expected: Tests pass (or same pass/fail as before).

- [ ] **Step 8: Commit**

```bash
git add web/components/layout/profile-dropdown.tsx web/components/posts/post-interaction-section.tsx web/components/posts/delete-post-button.tsx web/hooks/use-profile-controller.ts web/lib/notifications-api.ts
git commit -m "refactor: replace api.* calls with fetch + handleApiResponse in client components"
```

---

### Task 10: Update Server-Side Call Sites

**Files:**
- Modify: `web/app/posts/page.tsx`
- Modify: `web/app/posts/[id]/page.tsx`
- Modify: `web/app/[username]/page.tsx`
- Modify: `web/app/posts/create/actions.ts`

**Interfaces:**
- Consumes: `serverFetch` from `@/lib/server-fetch`, `ApiError` from `@/lib/errors`
- Removes dependency on `@/lib/api-server`

- [ ] **Step 1: Update `web/app/posts/page.tsx`**

Replace `import { apiServer } from "@/lib/api-server";` with `import { serverFetch } from "@/lib/server-fetch";` and `import { handleApiResponse, ApiError } from "@/lib/errors";`.

Replace getPosts function:
```ts
// Old:
async function getPosts() {
  try {
    const posts = await apiServer.get<Post[]>("/posts");
    return Array.isArray(posts) ? posts : [];
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      redirect("/login");
    }
    throw err;
  }
}
// New:
async function getPosts() {
  try {
    const res = await serverFetch("/api/posts");
    const posts = await handleApiResponse<Post[]>(res);
    return Array.isArray(posts) ? posts : [];
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      redirect("/login");
    }
    throw err;
  }
}
```

- [ ] **Step 2: Update `web/app/posts/[id]/page.tsx`**

Replace `import { apiServer } from "@/lib/api-server";` with `import { serverFetch } from "@/lib/server-fetch";` and `import { handleApiResponse, ApiError } from "@/lib/errors";`.

Replace getPost function:
```ts
// Old:
async function getPost(id: string) {
  try {
    const post = await apiServer.get<Post>(`/api/v1/posts/${id}`);
    if (!post) {
      notFound();
    }
    return post;
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 404) {
        notFound();
      }
    }
    throw err;
  }
}
// New:
async function getPost(id: string) {
  try {
    const res = await serverFetch(`/api/posts/${id}`);
    const post = await handleApiResponse<Post>(res);
    if (!post) {
      notFound();
    }
    return post;
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 404) {
        notFound();
      }
    }
    throw err;
  }
}
```

- [ ] **Step 3: Update `web/app/[username]/page.tsx`**

Replace `import { apiServer } from "@/lib/api-server";` with `import { serverFetch } from "@/lib/server-fetch";` and `import { handleApiResponse, ApiError } from "@/lib/errors";`.

Replace the parallel fetch block:
```ts
// Old:
    const [userData, followersData, followingData, postsData] = await Promise.all([
      apiServer.get<UserProfile>(`/users/${username}`),
      apiServer.get<FollowerInfo[]>(`/users/${username}/followers`),
      apiServer.get<FollowerInfo[]>(`/users/${username}/following`),
      apiServer.get<PostInfo[]>(`/users/${username}/posts`),
    ]);
// New:
    const [userData, followersData, followingData, postsData] = await Promise.all([
      serverFetch(`/api/users/${username}`).then(r => handleApiResponse<UserProfile>(r)),
      serverFetch(`/api/users/${username}/followers`).then(r => handleApiResponse<FollowerInfo[]>(r)),
      serverFetch(`/api/users/${username}/following`).then(r => handleApiResponse<FollowerInfo[]>(r)),
      serverFetch(`/api/users/${username}/posts`).then(r => handleApiResponse<PostInfo[]>(r)),
    ]);
```

- [ ] **Step 4: Update `web/app/posts/create/actions.ts`**

Replace `import { apiServer } from "@/lib/api-server";` with `import { serverFetch } from "@/lib/server-fetch";` and replace `import { ApiError } from "@/lib/api";` with `import { ApiError, handleApiResponse } from "@/lib/errors";`.

Replace the apiServer.post call:
```ts
// Old:
    await apiServer.post<CreatePostResponse>("/api/v1/posts", {
      title,
      content,
      tags: tagArray,
    });
// New:
    const res = await serverFetch("/api/posts", {
      method: "POST",
      body: JSON.stringify({ title, content, tags: tagArray }),
    });
    await handleApiResponse<CreatePostResponse>(res);
```

- [ ] **Step 5: Verify TypeScript compiles**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npx tsc --noEmit --pretty 2>&1 | head -30`
Expected: No errors related to these files.

- [ ] **Step 6: Run tests**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npm test -- --run 2>&1 | tail -20`
Expected: Tests pass.

- [ ] **Step 7: Commit**

```bash
git add web/app/posts/page.tsx web/app/posts/\[id\]/page.tsx web/app/\[username\]/page.tsx web/app/posts/create/actions.ts
git commit -m "refactor: replace apiServer calls with serverFetch in server components"
```

---

### Task 11: Remove Old Files and Clean Up

**Files:**
- Delete: `web/lib/api.ts`
- Delete: `web/lib/api-server.ts`
- Delete: `web/app/api/proxy/[...path]/route.ts`
- Delete: `web/app/api/proxy/[...path]/` directory
- Delete: `web/app/api/posts/schema.ts` (only imported by old route.ts being replaced)
- Delete: `web/app/api/posts/model.ts` (not imported by any file)
- Keep: `web/app/api/login/schema.ts` (still used by `login/page.tsx` for client-side Zod validation)

**Interfaces:**
- None produced — pure cleanup.

- [ ] **Step 1: Delete old files**

```bash
rm /Users/billykore/Kore/Golang/project1/web/lib/api.ts
rm /Users/billykore/Kore/Golang/project1/web/lib/api-server.ts
rm -rf /Users/billykore/Kore/Golang/project1/web/app/api/proxy/
rm /Users/billykore/Kore/Golang/project1/web/app/api/posts/schema.ts
rm /Users/billykore/Kore/Golang/project1/web/app/api/posts/model.ts
```

- [ ] **Step 2: Verify TypeScript compiles with no broken imports**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npx tsc --noEmit --pretty 2>&1 | head -30`
Expected: No errors from missing imports of deleted files.

- [ ] **Step 3: Run full test suite**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npm test -- --run 2>&1 | tail -20`
Expected: All tests pass.

- [ ] **Step 4: Commit**

```bash
git add -A web/lib/ web/app/api/
git commit -m "chore: remove old api.ts, api-server.ts, proxy route, and unused schemas"
```

---

### Task 12: Final Verification — Build Check

**Files:**
- None modified. Verification only.

- [ ] **Step 1: Run the Next.js build to verify all routes and imports resolve**

Run: `cd /Users/billykore/Kore/Golang/project1/web && npm run build 2>&1 | tail -40`
Expected: Build succeeds with no errors. API routes are compiled. Pages are rendered without errors.

- [ ] **Step 2: If build succeeds, mark complete**

No commit needed if build passes (Task 11 already committed cleanup).
