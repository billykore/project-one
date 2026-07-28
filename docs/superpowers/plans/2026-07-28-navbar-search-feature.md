# Navbar Search Feature — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a search bar to the navbar that submits to `/search?q=...` and displays results from the existing `GET /users/search` backend API.

**Architecture:** Navbar client component → navigate to `/search?q=...` → server component page fetches via Next.js API proxy route → Go backend `/users/search`. No autocomplete, no debounce, no pagination in v1.

**Tech Stack:** Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS 4

## Global Constraints

- No autocomplete, no suggestion dropdown, no debounce — bare search bar only
- Search bar lives in the existing `Navbar` client component
- Results page is a server component using `serverFetch`
- API call proxied through Next.js API route (existing `proxyToBackend` pattern)
- Types match the swagger `dto.SearchUsersResponse` shape
- Results page: simple list of `username` + `name`, each linking to `/[username]`
- Empty query: show message "Enter a search term"
- No results: show "No users found"
- Follow existing component patterns (Tailwind classes, `handleApiResponse`, `serverFetch`)

---

### Task 1: Add Search TypeScript Types

**Files:**
- Create: `web/lib/types/search.types.ts`

**Interfaces:**
- Produces: `SearchResult`, `SearchResponse` interfaces

- [ ] **Step 1: Create types file**

```typescript
export interface SearchResult {
  username: string;
  name: string;
}

export interface SearchResponse {
  data: SearchResult[];
  next_cursor: string;
  has_more: boolean;
}
```

- [ ] **Step 2: Verify TypeScript compiles**

```bash
cd web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/lib/types/search.types.ts
git commit -m "feat(search): add SearchResult and SearchResponse types"
```

---

### Task 2: Create API Proxy Route

**Files:**
- Create: `web/app/api/users/search/route.ts`

**Interfaces:**
- Consumes: `proxyToBackend` from `@/lib/api-proxy`
- Produces: Next.js API route `GET /api/users/search`

- [ ] **Step 1: Create the route handler**

```typescript
import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(req: Request) {
  const url = new URL(req.url);
  const q = url.searchParams.get("q") || "";
  const cursor = url.searchParams.get("cursor") || "";
  const limit = url.searchParams.get("limit") || "20";

  const params = new URLSearchParams({ q, limit });
  if (cursor) params.set("cursor", cursor);

  return proxyToBackend(req, `/users/search?${params.toString()}`);
}
```

- [ ] **Step 2: Verify TypeScript compiles**

```bash
cd web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/app/api/users/search/route.ts
git commit -m "feat(search): add /api/users/search proxy route"
```

---

### Task 3: Create Search Bar Component

**Files:**
- Create: `web/components/layout/search-bar.tsx`

**Interfaces:**
- Produces: `<SearchBar />` client component — controlled input, submits on Enter or button click, navigates to `/search?q=...`

- [ ] **Step 1: Create the search bar**

```tsx
"use client";

import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";

export default function SearchBar() {
  const [query, setQuery] = useState("");
  const router = useRouter();

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    const trimmed = query.trim();
    if (trimmed.length > 0) {
      router.push(`/search?q=${encodeURIComponent(trimmed)}`);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex items-center">
      <div className="relative">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search users..."
          aria-label="Search users"
          className="w-48 rounded-lg border border-gray-300 bg-gray-50 py-1.5 pl-9 pr-3 text-sm text-gray-900 placeholder-gray-400 transition-all focus:border-indigo-400 focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-200 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500 dark:focus:border-indigo-500 dark:focus:ring-indigo-800 sm:w-56"
        />
        <button
          type="submit"
          aria-label="Search"
          className="absolute left-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-indigo-500 dark:hover:text-indigo-400"
        >
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
          </svg>
        </button>
      </div>
    </form>
  );
}
```

- [ ] **Step 2: Verify TypeScript compiles**

```bash
cd web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/components/layout/search-bar.tsx
git commit -m "feat(search): add SearchBar component with router navigation"
```

---

### Task 4: Integrate Search Bar into Navbar

**Files:**
- Modify: `web/components/layout/navbar.tsx:1-5` (add import)
- Modify: `web/components/layout/navbar.tsx:85-90` (insert component between left/right sections)

**Interfaces:**
- Consumes: `SearchBar` from `@/components/layout/search-bar`
- Produces: Updated `<Navbar>` with search bar

- [ ] **Step 1: Add import**

Add after the existing `import ProfileDropdown...` line:

```tsx
import SearchBar from "@/components/layout/search-bar";
```

- [ ] **Step 2: Insert SearchBar in the center area**

Replace the spacer between left and right sections. In the navbar, between the left logo/title section and the right actions section, add:

```tsx
      {/* Center: Search bar */}
      <div className="flex-1 flex justify-center mx-4">
        <SearchBar />
      </div>
```

This goes right before the right-side `{/* Right side: Nav links and dropdowns */}` section. The full structure becomes:

```
<nav>
  {/* Left side: Logo & page title */}
  {/* Center: Search bar */}
  {/* Right side: Nav links and dropdowns */}
</nav>
```

- [ ] **Step 3: Adjust navbar flex to accommodate three sections**

The nav currently uses `flex items-center justify-between`. Change `justify-between` to keep three-section layout:

No change needed — `justify-between` with the new center div will naturally create three sections since the left and right divs have natural widths and the center will flex-grow.

- [ ] **Step 4: Verify TypeScript compiles**

```bash
cd web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add web/components/layout/navbar.tsx
git commit -m "feat(search): integrate SearchBar into navbar center"
```

---

### Task 5: Create Search Results Page

**Files:**
- Create: `web/app/search/page.tsx`

**Interfaces:**
- Consumes: `SearchResult`, `SearchResponse` types, `serverFetch`, `handleApiResponse`
- Produces: Search results page at `/search?q=...`

- [ ] **Step 1: Create the search results page**

```tsx
import Link from "next/link";
import { serverFetch } from "@/lib/server-fetch";
import { handleApiResponse } from "@/lib/errors";
import type { SearchResponse } from "@/lib/types/search.types";

interface SearchPageProps {
  searchParams: Promise<{ q?: string }>;
}

export default async function SearchPage({ searchParams }: SearchPageProps) {
  const { q } = await searchParams;

  if (!q || q.trim().length === 0) {
    return (
      <main className="mx-auto max-w-2xl px-4 py-12">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">Search Users</h1>
        <p className="mt-4 text-gray-500 dark:text-gray-400">Enter a search term to find users.</p>
      </main>
    );
  }

  const res = await serverFetch(`/api/users/search?q=${encodeURIComponent(q)}&limit=20`);

  if (!res.ok) {
    // ponytail: show generic error; add detailed error states if users complain
    return (
      <main className="mx-auto max-w-2xl px-4 py-12">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">
          Search results for &ldquo;{q}&rdquo;
        </h1>
        <p className="mt-4 text-gray-500 dark:text-gray-400">
          Something went wrong. Please try again.
        </p>
      </main>
    );
  }

  const { data: results } = await handleApiResponse<SearchResponse>(res);

  return (
    <main className="mx-auto max-w-2xl px-4 py-12">
      <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">
        Search results for &ldquo;{q}&rdquo;
      </h1>

      {results.length === 0 ? (
        <p className="mt-4 text-gray-500 dark:text-gray-400">
          No users found matching &ldquo;{q}&rdquo;.
        </p>
      ) : (
        <ul className="mt-6 divide-y divide-gray-200 dark:divide-gray-800">
          {results.map((user) => (
            <li key={user.username}>
              <Link
                href={`/${user.username}`}
                className="flex items-center gap-3 py-3 transition-colors hover:bg-gray-50 dark:hover:bg-gray-900 rounded-lg px-2 -mx-2"
              >
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-linear-to-br from-indigo-500 to-purple-600 text-sm font-medium text-white">
                  {user.name.charAt(0).toUpperCase()}
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-900 dark:text-white">
                    {user.name}
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    @{user.username}
                  </p>
                </div>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}
```

- [ ] **Step 2: Verify TypeScript compiles**

```bash
cd web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Run frontend dev server to verify**

```bash
cd web && npm run dev
```
Visit `http://localhost:3000/search?q=bil` — verify results render.

- [ ] **Step 4: Commit**

```bash
git add web/app/search/page.tsx
git commit -m "feat(search): add search results page at /search"
```

---

### Task 6: Final Verification

- [ ] **Step 1: Run TypeScript check**

```bash
cd web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 2: Run existing tests**

```bash
cd web && npm test
```
Expected: all existing tests pass.

- [ ] **Step 3: Manual smoke test**

Start both backend (`make run`) and frontend (`npm run dev`):

1. Visit `http://localhost:3000` — search bar visible in navbar
2. Type `bil` → press Enter → navigates to `/search?q=bil`
3. Results displayed with name, username, initials avatar
4. Click a result → navigates to user profile
5. Empty search → "Enter a search term" message
6. No-match query → "No users found" message
