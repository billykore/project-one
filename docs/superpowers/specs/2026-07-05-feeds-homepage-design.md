# Feeds Homepage — Design Spec

> **Created:** 2026-07-05 | **Status:** Approved

## Overview

Replace the current dashboard homepage (`app/page.tsx`) with a feed view showing posts from the authenticated user and their followed users, fetched from the `GET /feeds` API endpoint with infinite scroll.

## Architecture

**Hybrid rendering**: Server Component fetches the first page (SSR for fast initial paint), then hands off to a Client Component for interactive infinite scroll.

```
app/page.tsx (Server Component)
  └─ serverFetch("/api/feeds?limit=10")
  └─ renders <FeedView initialPosts={...} ... />

components/posts/FeedView.tsx (Client Component)
  └─ IntersectionObserver sentinel
  └─ fetch("/api/feeds?cursor=...&limit=10") on intersect
  └─ appends posts, tracks cursor/hasMore/loading/error state

app/api/feeds/route.ts (API Route)
  └─ proxyToBackend → GET localhost:8080/feeds
  └─ forwards Set-Cookie for token refresh
```

## Components

### `app/page.tsx` — Homepage (modified)

- Changed from Client to Server Component
- `serverFetch("/api/feeds?limit=10")` for initial data
- Renders `Navbar` + `FeedView` with initial props
- On 401: `redirect("/login")`
- On other errors: renders `ErrorDisplay`

### `components/posts/FeedView.tsx` — Feed View (new)

Client Component with:

**Props:**
- `initialPosts: Post[]`
- `nextCursor: string | null`
- `hasMore: boolean`

**State:**
- `posts: Post[]` (seeded from `initialPosts`)
- `cursor: string | null` (next page cursor)
- `hasMore: boolean`
- `isLoading: boolean`
- `error: string | null`

**Behavior:**
- Renders vertical single-column list of post cards
- Sentinel `<div>` at bottom observed via `IntersectionObserver`
- On intersect: fetches next page, appends results, deduplicates by `id`
- Guards against concurrent fetches with `isFetchingRef`

**States covered:**
- **Loading:** skeleton spinner at bottom while fetching next page
- **Empty:** "No posts yet" message with "Create Post" and "Browse Posts" links
- **Error:** inline error banner with "Retry" button
- **End of feed:** subtle "You're all caught up" message when `hasMore === false`

### `app/api/feeds/route.ts` — API Route (new)

- Proxies `GET` requests to backend `/feeds` with query params
- Forwards cookies for auth
- Returns JSON response with `Set-Cookie` headers for token refresh

## Data Types

Reusing existing `Post` type from `lib/types/post.types.ts`:

```typescript
interface Post {
  id: number;
  title: string;
  content: string;
  tags?: string[];
  author?: string;
  created_at: string;
  updated_at: string;
  like_count?: number;
}
```

API response shape (from `dto.FeedResponse`):
```typescript
interface FeedResponse {
  data: Post[];
  has_more: boolean;
  next_cursor: string;
}
```

## Post Card Design

Vertical stacked cards (single column, max-w-2xl):

```
┌─────────────────────────────────────┐
│ 👤 Author Name          📅 Date     │
│                                     │
│ Post Title (bold, link to detail)   │
│ Content preview (150 chars max)...  │
│                                     │
│ 🏷 tag1 tag2    ❤️ 12 likes    →   │
└─────────────────────────────────────┘
```

- Author name links to `/[username]`
- Title links to `/posts/[id]`
- Tags rendered as small pills
- Like count shown as icon + number
- Hover: subtle shadow lift + ring color transition (consistent with existing cards)

## Error Handling

| Scenario | Behavior |
|----------|----------|
| 401 Unauthorized | Redirect to `/login` |
| Network error on initial load | `ErrorDisplay` with retry + home link |
| Network error on "load more" | Inline error banner, existing posts remain, retry button |
| Empty feed (0 posts) | "No posts yet" empty state with CTAs |
| Duplicate posts from pagination | Deduplicate by `id` when appending |
| Rapid scroll / double fetch | `isFetchingRef` guards concurrent requests |

## Files Changed/Created

| File | Action | Description |
|------|--------|-------------|
| `web/app/page.tsx` | Modify | Replace dashboard with feed (Server Component) |
| `web/components/posts/FeedView.tsx` | Create | Client Component with infinite scroll |
| `web/app/api/feeds/route.ts` | Create | API proxy route for `/feeds` |
| `web/lib/types/post.types.ts` | No change | `Post` type already matches `PostItem` |

## Testing Strategy

- **FeedView unit tests:** renders with posts, empty state, error state, infinite scroll (mock IntersectionObserver)
- **API route integration test:** proxies correctly, forwards cookies, handles backend errors
- **Homepage integration test:** SSR renders initial posts, 401 redirects, error fallback

## Dependencies

- None new. Uses existing: React 19, Next.js 16, Tailwind CSS 4, IntersectionObserver (native browser API).
