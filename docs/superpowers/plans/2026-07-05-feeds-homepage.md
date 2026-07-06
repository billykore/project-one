# Feeds Homepage — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the dashboard homepage with a feed of posts from followed users, with server-rendered initial page and client-side infinite scroll.

**Architecture:** Server Component (`app/page.tsx`) fetches page 1 via `serverFetch`, renders a new `FeedView` Client Component. `FeedView` uses `IntersectionObserver` to trigger cursor-paginated fetches through a new `/api/feeds` proxy route.

**Tech Stack:** Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS 4, Vitest + jsdom

## Global Constraints

- Next.js 16.2.9, React 19.2.4, Tailwind CSS 4
- Tests use Vitest + jsdom, follow existing `createRoot` + `act()` pattern
- API routes use `proxyToBackend` from `@/lib/api-proxy`
- TypeScript strict mode
- No new dependencies

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `web/app/api/feeds/route.ts` | Create | Proxy `GET /feeds` to backend |
| `web/components/posts/FeedView.tsx` | Create | Client component: infinite scroll feed |
| `web/tests/components/feed-view.test.tsx` | Create | Unit tests for FeedView |
| `web/app/page.tsx` | Modify | Server component: fetch + render FeedView |

---

### Task 1: Create the `/api/feeds` proxy route

**Files:**
- Create: `web/app/api/feeds/route.ts`

**Interfaces:**
- Consumes: `proxyToBackend` from `@/lib/api-proxy`
- Produces: `GET` handler that proxies to backend `/feeds`

- [ ] **Step 1: Create the route file**

```typescript
import { proxyToBackend } from "@/lib/api-proxy";

export async function GET(req: Request) {
  return proxyToBackend(req, "/feeds");
}
```

- [ ] **Step 2: Verify the route compiles**

Run: `cd web && npx tsc --noEmit app/api/feeds/route.ts`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/app/api/feeds/route.ts
git commit -m "feat: add /api/feeds proxy route"
```

---

### Task 2: Write unit tests for FeedView

**Files:**
- Create: `web/tests/components/feed-view.test.tsx`

**Interfaces:**
- Consumes: `FeedView` props (`initialPosts: Post[]`, `nextCursor: string | null`, `hasMore: boolean`)
- Produces: Test coverage for all states

- [ ] **Step 1: Write the test file with all state tests**

```typescript
import React from "react";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { FeedView } from "@/components/posts/FeedView";
import type { Post } from "@/lib/types/post.types";

const mockPost = (id: number): Post => ({
  id,
  title: `Post ${id}`,
  content: `Content for post ${id}`,
  author: `author${id}`,
  tags: ["tag1", "tag2"],
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  like_count: id * 3,
});

// Mock IntersectionObserver
beforeEach(() => {
  vi.stubGlobal(
    "IntersectionObserver",
    vi.fn(() => ({
      observe: vi.fn(),
      unobserve: vi.fn(),
      disconnect: vi.fn(),
    }))
  );
});

describe("FeedView", () => {
  it("renders a list of post cards from initialPosts", async () => {
    const posts = [mockPost(1), mockPost(2)];
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <FeedView initialPosts={posts} nextCursor="cursor-1" hasMore={true} />
      );
    });

    expect(container.textContent).toContain("Post 1");
    expect(container.textContent).toContain("Post 2");
    expect(container.textContent).toContain("Content for post 1");
    expect(container.textContent).toContain("Content for post 2");
  });

  it("shows empty state when no posts and no more pages", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <FeedView initialPosts={[]} nextCursor={null} hasMore={false} />
      );
    });

    expect(container.textContent).toContain("No posts yet");
    expect(container.textContent).toContain("Create Post");
  });

  it("shows author names on cards", async () => {
    const posts = [mockPost(1)];
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <FeedView initialPosts={posts} nextCursor={null} hasMore={false} />
      );
    });

    expect(container.textContent).toContain("author1");
  });

  it("shows like count on cards", async () => {
    const posts = [mockPost(1)];
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <FeedView initialPosts={posts} nextCursor={null} hasMore={false} />
      );
    });

    expect(container.textContent).toContain("3");
  });

  it("renders sentinel div for IntersectionObserver", async () => {
    const posts = [mockPost(1)];
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <FeedView initialPosts={posts} nextCursor="cursor-1" hasMore={true} />
      );
    });

    // The sentinel should exist in the DOM
    const sentinelDivs = Array.from(container.querySelectorAll("div")).filter(
      (d) => d.getAttribute("data-feed-sentinel") !== null
    );
    expect(sentinelDivs.length).toBe(1);
  });

  it("shows 'no more posts' when hasMore is false and posts exist", async () => {
    const posts = [mockPost(1)];
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <FeedView initialPosts={posts} nextCursor={null} hasMore={false} />
      );
    });

    expect(container.textContent).toContain("You're all caught up");
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd web && npx vitest run tests/components/feed-view.test.tsx`
Expected: FAIL — `FeedView` module not found

- [ ] **Step 3: Commit**

```bash
git add web/tests/components/feed-view.test.tsx
git commit -m "test: add FeedView unit tests"
```

---

### Task 3: Create the FeedView component

**Files:**
- Create: `web/components/posts/FeedView.tsx`

**Interfaces:**
- Consumes: `Post` from `@/lib/types/post.types`
- Produces: `FeedView` default export with props `initialPosts`, `nextCursor`, `hasMore`

- [ ] **Step 1: Create the FeedView component**

```typescript
"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import Link from "next/link";
import type { Post } from "@/lib/types/post.types";

interface FeedViewProps {
  initialPosts: Post[];
  nextCursor: string | null;
  hasMore: boolean;
}

const TRUNCATE_LIMIT = 150;

function truncateContent(content: string): string {
  if (content.length <= TRUNCATE_LIMIT) return content;
  return content.slice(0, TRUNCATE_LIMIT) + "...";
}

export function FeedView({ initialPosts, nextCursor, hasMore }: FeedViewProps) {
  const [posts, setPosts] = useState<Post[]>(initialPosts);
  const [cursor, setCursor] = useState<string | null>(nextCursor);
  const [moreAvailable, setMoreAvailable] = useState(hasMore);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isFetchingRef = useRef(false);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  const fetchNextPage = useCallback(async () => {
    if (isFetchingRef.current || !moreAvailable || cursor === null) return;
    isFetchingRef.current = true;
    setIsLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams({ limit: "10", cursor });
      const res = await fetch(`/api/feeds?${params.toString()}`);
      if (res.status === 401) {
        window.location.href = "/login";
        return;
      }
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error || `Request failed (${res.status})`);
      }
      const data = await res.json();
      const incoming: Post[] = data.data ?? [];

      setPosts((prev) => {
        const existingIds = new Set(prev.map((p) => p.id));
        const newPosts = incoming.filter((p) => !existingIds.has(p.id));
        return [...prev, ...newPosts];
      });
      setCursor(data.next_cursor ?? null);
      setMoreAvailable(data.has_more ?? false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load more posts");
    } finally {
      setIsLoading(false);
      isFetchingRef.current = false;
    }
  }, [cursor, moreAvailable]);

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          fetchNextPage();
        }
      },
      { rootMargin: "200px" }
    );

    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [fetchNextPage]);

  // Empty state
  if (posts.length === 0 && !moreAvailable && !isLoading) {
    return (
      <div className="flex flex-col items-center justify-center rounded-2xl bg-white p-12 shadow-sm ring-1 ring-gray-200 dark:bg-gray-900 dark:ring-gray-800">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-gray-800 mb-4">
          <svg className="h-8 w-8 text-gray-400" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
          </svg>
        </div>
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white">No posts yet</h2>
        <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
          Follow some users or create your first post to populate your feed.
        </p>
        <div className="mt-6 flex gap-3">
          <Link
            href="/posts/create"
            className="inline-flex items-center gap-2 rounded-lg bg-linear-to-r from-indigo-600 to-indigo-700 px-5 py-2.5 text-sm font-medium text-white shadow-sm transition-all hover:from-indigo-500 hover:to-indigo-600 hover:shadow-md active:scale-[0.98]"
          >
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
            </svg>
            Create Post
          </Link>
          <Link
            href="/posts"
            className="inline-flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-5 py-2.5 text-sm font-medium text-gray-700 shadow-sm transition-all hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
          >
            Browse Posts
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {posts.map((post) => (
        <Link
          key={post.id}
          href={`/posts/${post.id}`}
          className="group block rounded-2xl bg-white p-6 shadow-sm ring-1 ring-gray-200 transition-all hover:shadow-lg hover:ring-gray-300 hover:-translate-y-0.5 dark:bg-gray-900 dark:ring-gray-800 dark:hover:ring-gray-700"
        >
          {/* Author + Date row */}
          <div className="flex items-center justify-between mb-3">
            {post.author ? (
              <Link
                href={`/${post.author}`}
                onClick={(e) => e.stopPropagation()}
                className="text-sm font-medium text-gray-700 hover:text-indigo-600 dark:text-gray-300 dark:hover:text-indigo-400 transition-colors"
              >
                {post.author}
              </Link>
            ) : (
              <span className="text-sm text-gray-400 dark:text-gray-500">Unknown</span>
            )}
            <span className="text-xs text-gray-400 dark:text-gray-500">
              {new Date(post.created_at).toLocaleDateString(undefined, {
                month: "short",
                day: "numeric",
                year: "numeric",
              })}
            </span>
          </div>

          {/* Title */}
          <h3 className="text-lg font-semibold tracking-tight text-gray-900 group-hover:text-indigo-600 dark:text-white dark:group-hover:text-indigo-400 transition-colors">
            {post.title}
          </h3>

          {/* Content preview */}
          <p className="mt-2 text-sm leading-relaxed text-gray-500 dark:text-gray-400">
            {truncateContent(post.content)}
          </p>

          {/* Tags + Like count + Read more */}
          <div className="mt-4 flex items-center justify-between">
            <div className="flex items-center gap-2 flex-wrap">
              {post.tags?.map((tag) => (
                <span
                  key={tag}
                  className="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-600 dark:bg-gray-800 dark:text-gray-400"
                >
                  {tag}
                </span>
              ))}
            </div>
            <div className="flex items-center gap-3 text-xs">
              {post.like_count !== undefined && (
                <span className="flex items-center gap-1 text-gray-400 dark:text-gray-500">
                  <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12z" />
                  </svg>
                  {post.like_count}
                </span>
              )}
              <span className="font-medium text-indigo-600 dark:text-indigo-400 group-hover:translate-x-0.5 transition-transform">
                Read more →
              </span>
            </div>
          </div>
        </Link>
      ))}

      {/* Sentinel for infinite scroll */}
      {moreAvailable && (
        <div ref={sentinelRef} data-feed-sentinel="true" className="h-4" />
      )}

      {/* Loading indicator */}
      {isLoading && (
        <div className="flex items-center justify-center gap-3 py-6 text-gray-500 dark:text-gray-400">
          <svg className="h-5 w-5 animate-spin" viewBox="0 0 24 24" fill="none">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          <span className="text-sm font-medium">Loading more posts…</span>
        </div>
      )}

      {/* Error with retry */}
      {error && (
        <div className="flex items-center justify-between rounded-xl bg-red-50 px-5 py-4 ring-1 ring-red-200 dark:bg-red-950/30 dark:ring-red-800">
          <p className="text-sm text-red-700 dark:text-red-400">{error}</p>
          <button
            onClick={fetchNextPage}
            className="rounded-lg bg-red-100 px-3 py-1.5 text-sm font-medium text-red-700 hover:bg-red-200 dark:bg-red-900/50 dark:text-red-300 dark:hover:bg-red-900 transition-colors"
          >
            Retry
          </button>
        </div>
      )}

      {/* End of feed */}
      {!moreAvailable && posts.length > 0 && !isLoading && (
        <div className="py-8 text-center">
          <p className="text-sm text-gray-400 dark:text-gray-500">You&apos;re all caught up ✨</p>
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Run tests to verify they pass**

Run: `cd web && npx vitest run tests/components/feed-view.test.tsx`
Expected: All 6 tests PASS

- [ ] **Step 3: Commit**

```bash
git add web/components/posts/FeedView.tsx
git commit -m "feat: add FeedView component with infinite scroll"
```

---

### Task 4: Update homepage to render the feed

**Files:**
- Modify: `web/app/page.tsx`

**Interfaces:**
- Consumes: `serverFetch` from `@/lib/server-fetch`, `FeedView` from `@/components/posts/FeedView`
- Produces: Server Component that fetches feeds and renders FeedView

- [ ] **Step 1: Rewrite page.tsx as a Server Component**

```typescript
import { redirect } from "next/navigation";
import { serverFetch } from "@/lib/server-fetch";
import { ApiError, handleApiResponse } from "@/lib/errors";
import { FeedView } from "@/components/posts/FeedView";
import Navbar from "@/components/layout/navbar";
import type { Post } from "@/lib/types/post.types";

interface FeedApiResponse {
  data: Post[];
  has_more: boolean;
  next_cursor: string;
}

async function getFeed(): Promise<FeedApiResponse> {
  try {
    const res = await serverFetch("/api/feeds?limit=10");
    return await handleApiResponse<FeedApiResponse>(res);
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      redirect("/login");
    }
    throw err;
  }
}

export default async function HomePage() {
  let feed: FeedApiResponse;
  try {
    feed = await getFeed();
  } catch {
    return (
      <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-gray-950">
        <Navbar pageTitle="Home" />
        <main className="flex flex-1 items-center justify-center p-6">
          <div className="flex flex-col items-center gap-4 text-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30">
              <svg className="h-8 w-8 text-red-500" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white">Failed to load feed</h2>
            <p className="text-sm text-gray-500 dark:text-gray-400">Something went wrong. Please try again.</p>
            <a
              href="/"
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 transition-colors"
            >
              Retry
            </a>
          </div>
        </main>
        <footer className="py-6 text-center text-xs text-gray-400 dark:text-gray-600">
          &copy; {new Date().getFullYear()} Project One. All rights reserved.
        </footer>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-gray-950">
      <Navbar pageTitle="Home" />

      <main className="mx-auto w-full max-w-2xl flex-1 p-6 sm:p-8">
        <FeedView
          initialPosts={feed.data}
          nextCursor={feed.next_cursor}
          hasMore={feed.has_more}
        />
      </main>

      <footer className="py-6 text-center text-xs text-gray-400 dark:text-gray-600">
        &copy; {new Date().getFullYear()} Project One. All rights reserved.
      </footer>
    </div>
  );
}
```

- [ ] **Step 2: Run the full test suite**

Run: `cd web && npx vitest run`
Expected: All existing tests + new FeedView tests PASS

- [ ] **Step 3: Verify TypeScript compilation**

Run: `cd web && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add web/app/page.tsx
git commit -m "feat: replace dashboard with feeds homepage"
```

---

### Task 5: Final verification

- [ ] **Step 1: Run the full test suite one final time**

Run: `cd web && npx vitest run`
Expected: All tests PASS, no regressions

- [ ] **Step 2: Build check**

Run: `cd web && npm run build`
Expected: Successful build with no errors
