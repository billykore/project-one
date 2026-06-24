# Login API Refactor — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the login page to use the `/api/login` Next.js API route instead of calling the Go backend directly, and remove the inline Zod schema duplication.

**Architecture:** The `route.ts` becomes a transparent proxy — validates request body with the shared `loginRequestBodySchema`, forwards to the Go backend at `{API_URL}/api/v1/auth/login`, and relays the response (status code + JSON body) back to the client. The login page imports the shared schema and calls `/api/login` via `fetch` instead of `api.post("/api/v1/auth/login", ...)`.

**Tech Stack:** Next.js 16 (App Router), React 19, TypeScript, Zod, Vitest

## Global Constraints

- Zod validation schema lives in `web/app/(auth)/api/login/schema.ts` (already exists) — no changes to it
- API route uses `process.env.API_URL` for the backend base URL (documented in `web/README.md`)
- Login page is a client component (`"use client"`) — `fetch` calls go through the browser
- Existing error UX is preserved: 400/401 → inline error, 500/network → error modal
- Test files go in `web/tests/` with `.test.ts` or `.test.tsx` extension
- Vitest v4, jsdom environment

---

### Task 1: Implement Route Transparent Proxy

**Files:**
- Modify: `web/app/(auth)/api/login/route.ts`

**Interfaces:**
- Consumes: `loginRequestBodySchema`, `LoginRequestBody` from `./schema` (already imported)
- Produces: `POST(req: Request) => Promise<Response>` — validates body, proxies to backend, relays response

- [ ] **Step 1: Replace the empty route with the proxy implementation**

Replace the current skeleton:

```ts
import { LoginRequestBody, loginRequestBodySchema } from "./schema";

export async function POST(req: Request) {
    
}
```

With:

```ts
import { loginRequestBodySchema } from "./schema";

export async function POST(req: Request) {
  let body: unknown;
  try {
    body = await req.json();
  } catch {
    return Response.json({ error: "Invalid JSON body" }, { status: 400 });
  }

  const parsed = loginRequestBodySchema.safeParse(body);
  if (!parsed.success) {
    return Response.json(
      { error: parsed.error.issues[0]?.message || "Invalid request" },
      { status: 400 }
    );
  }

  const backendUrl = `${process.env.API_URL || "http://localhost:8080"}/api/v1/auth/login`;

  let backend: Response;
  try {
    backend = await fetch(backendUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(parsed.data),
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

  return Response.json(data, { status: backend.status });
}
```

- [ ] **Step 2: Verify the file compiles**

Run: `cd web && npx tsc --noEmit --pretty app/\(auth\)/api/login/route.ts`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/app/\(auth\)/api/login/route.ts
git commit -m "feat: implement login API route transparent proxy"
```

---

### Task 2: Refactor Login Page

**Files:**
- Modify: `web/app/(auth)/login/page.tsx`
- Create: `web/tests/pages/login-page.test.tsx`

**Interfaces:**
- Consumes: `loginRequestBodySchema`, `LoginRequestBody` from `@/app/(auth)/api/login/schema`
- Produces: Same LoginPage component, but fetches `/api/login` instead of `/api/v1/auth/login`

- [ ] **Step 1: Write the failing test**

```ts
import React from 'react';
import { act } from 'react';
import { createRoot } from 'react-dom/client';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

// Mock next/navigation
const mockPush = vi.fn();
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
}));

// Mock next/link
vi.mock('next/link', () => ({
  default: ({ children, ...props }: { children: React.ReactNode; [key: string]: unknown }) =>
    React.createElement('a', props, children),
}));

// Mock the error modal hook
const mockShowError = vi.fn();
vi.mock('@/hooks/use-error-modal', () => ({
  useErrorModal: () => ({ showError: mockShowError }),
}));

let LoginPage: React.ComponentType;
let container: HTMLDivElement;

describe('LoginPage', () => {
  beforeEach(async () => {
    mockPush.mockReset();
    mockShowError.mockReset();
    global.fetch = vi.fn();
    container = document.createElement('div');
    document.body.appendChild(container);

    const mod = await import('@/app/(auth)/login/page');
    LoginPage = mod.default;
  });

  afterEach(() => {
    document.body.removeChild(container);
    vi.restoreAllMocks();
  });

  it('calls /api/login on form submit', async () => {
    vi.mocked(global.fetch).mockResolvedValue(
      new Response(JSON.stringify({ message: "ok", username: "testuser" }), { status: 200 })
    );

    const root = createRoot(container);
    await act(async () => {
      root.render(React.createElement(LoginPage));
    });

    // Fill in the form
    const emailInput = container.querySelector('#email') as HTMLInputElement;
    const passwordInput = container.querySelector('#password') as HTMLInputElement;
    const form = container.querySelector('form')!;

    await act(async () => {
      emailInput.value = 'test@example.com';
      emailInput.dispatchEvent(new Event('change', { bubbles: true }));
      passwordInput.value = 'password123';
      passwordInput.dispatchEvent(new Event('change', { bubbles: true }));
      form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    });

    expect(global.fetch).toHaveBeenCalledWith(
      '/api/login',
      expect.objectContaining({
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      })
    );

    const callArgs = vi.mocked(global.fetch).mock.calls[0];
    const body = JSON.parse(callArgs[1]!.body as string);
    expect(body).toEqual({ email: 'test@example.com', password: 'password123' });
  });

  it('redirects to / on successful login', async () => {
    vi.mocked(global.fetch).mockResolvedValue(
      new Response(JSON.stringify({ message: "ok", username: "testuser" }), { status: 200 })
    );

    const root = createRoot(container);
    await act(async () => {
      root.render(React.createElement(LoginPage));
    });

    const emailInput = container.querySelector('#email') as HTMLInputElement;
    const passwordInput = container.querySelector('#password') as HTMLInputElement;
    const form = container.querySelector('form')!;

    await act(async () => {
      emailInput.value = 'test@example.com';
      emailInput.dispatchEvent(new Event('change', { bubbles: true }));
      passwordInput.value = 'password123';
      passwordInput.dispatchEvent(new Event('change', { bubbles: true }));
      form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    });

    expect(mockPush).toHaveBeenCalledWith('/');
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npx vitest run tests/pages/login-page.test.tsx`
Expected: FAIL — the test asserts `fetch` is called with `/api/login`, but the current page calls the backend directly via `api.post`

- [ ] **Step 3: Refactor the login page**

Remove the inline Zod schema and `api` import. Replace the submit handler to call `/api/login`.

Replace lines 1-8 (imports):

```tsx
"use client";

import React, { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { z, ZodError } from "zod";
import { api, ApiError } from "@/lib/api";
import { InputField } from "@/components/ui/input";
import { useErrorModal } from "@/hooks/use-error-modal";
```

With:

```tsx
"use client";

import React, { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ZodError } from "zod";
import { ApiError } from "@/lib/api";
import { InputField } from "@/components/ui/input";
import { useErrorModal } from "@/hooks/use-error-modal";
import { loginRequestBodySchema, LoginRequestBody } from "@/app/(auth)/api/login/schema";
```

Replace the inline schema block (lines 11-18):

```tsx
// ponytail: inline login schema and validation types
const loginSchema = z.object({
  email: z.string().min(1, "Email is required").pipe(z.email("Please enter a valid email address")),
  password: z
    .string()
    .min(1, "Password is required")
    .min(8, "Password must be at least 8 characters long"),
});

type LoginFormData = z.infer<typeof loginSchema>;
```

With:

```tsx
type LoginFormData = LoginRequestBody;
```

Replace the `validate` function's call to `loginSchema.parse`:

In the `validate` function, replace:
```tsx
      loginSchema.parse(formData);
```

With:
```tsx
      loginRequestBodySchema.parse(formData);
```

Replace the submit handler in `handleSubmit` — the try block currently reads:

```tsx
    try {
      await api.post<LoginResponse>("/api/v1/auth/login", {
        email: formData.email.trim(),
        password: formData.password,
      });
      router.push("/");
    } catch (err) {
      if (err instanceof ApiError && (err.status === 401 || err.status === 400)) {
        setErrors({ general: err.message });
        return;
      }
      const errorMessage = err instanceof Error ? err.message : "An error occurred. Please try again later.";
      showError(errorMessage);
    } finally {
```

Replace with:

```tsx
    try {
      const res = await fetch("/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email: formData.email.trim(),
          password: formData.password,
        }),
      });

      if (!res.ok) {
        const errBody = await res.json().catch(() => ({ error: "Login failed" }));
        throw new ApiError(errBody.error || "Login failed", res.status);
      }

      const data: LoginResponse = await res.json();
      router.push("/");
    } catch (err) {
      if (err instanceof ApiError && (err.status === 401 || err.status === 400)) {
        setErrors({ general: err.message });
        return;
      }
      const errorMessage = err instanceof Error ? err.message : "An error occurred. Please try again later.";
      showError(errorMessage);
    } finally {
```

Remove the unused `LoginResponse` interface if it's no longer needed — wait, it IS still used for `const data: LoginResponse`. Keep it.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd web && npx vitest run tests/pages/login-page.test.tsx`
Expected: 2 tests PASS

- [ ] **Step 5: Run full test suite to check for regressions**

Run: `cd web && npx vitest run`
Expected: All existing tests still pass

- [ ] **Step 6: Verify TypeScript compiles**

Run: `cd web && npx tsc --noEmit --pretty`
Expected: No errors

- [ ] **Step 7: Commit**

```bash
git add web/app/\(auth\)/login/page.tsx web/tests/pages/login-page.test.tsx
git commit -m "refactor: login page uses /api/login route with shared schema"
```
