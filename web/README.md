# Project One — Frontend

A full-featured social media frontend built with **Next.js 16** (App Router), **React 19**, and **Tailwind CSS 4**, connecting to a Go backend via REST and WebSocket.

---

## ✨ Features

### Authentication

- **Login page** (`/login`) — email/password form with Zod validation.
- **JWT-based auth** — `access_token` HttpOnly cookie set by the backend.
- **Session tracking** via `localStorage` username.
- **Logout** with confirmation modal, API call, and redirect.
- **Guest mode** — read-only access to public content; like/comment actions gracefully disabled.

### Post Management

- **Create posts** (`/posts/create`) — server action form with title, content, and tags (comma-separated).
- **View all posts** (`/posts`) — server-rendered responsive grid (1/2/3 columns), 401 redirect.
- **Post detail** (`/posts/[id]`) — full content, tags, author metadata, like count, and comment thread.
- **Like/unlike** — optimistic updates with rollback on API failure; guest-disabled with tooltip.
- **Delete posts** — author-only delete with confirmation modal, focus management, escape key support.
- **Comments** — add, inline edit (author only), and delete with confirmation.

### User Profiles

- **Dynamic routes** (`/[username]`) — server-rendered profile pages.
- **Profile header** — gradient background, initials avatar, user details.
- **Email masking** — email shown only to the profile owner.
- **Follow/unfollow** — optimistic UI toggle with instant feedback.
- **Follower/Following modals** — clickable counts, user list with mutual badges.
- **Change password** form (owner only) with Zod validation.

### Real-time Notifications

- **WebSocket client** with JWT token auth (query param fallback for browser WS).
- **Capped exponential backoff** reconnect (1s → 30s, max 10 attempts).
- **REST fallback** — fetches missed notifications on reconnect.
- **Deduplication** — by notification `id` to avoid duplicates.
- **Optimistic** mark-as-read / mark-all-as-read with rollback on error.
- **Connection indicator** — Live / Reconnecting / Offline states.
- **Notification types** — follow, like, and comment events.

### Navigation

- **Responsive Navbar** — logo, dynamic page title, Create Post button.
- **NotificationDropdown** — bell icon with unread badge, live indicator, scrollable list.
- **ProfileDropdown** — initials avatar (gradient), links to Dashboard / Profile / Posts / Create Post, logout.

### UI/UX

- **Tailwind CSS v4** with light/dark theme support.
- **Skeleton loading states** on the posts grid.
- **Error boundaries** on posts pages; dedicated `/error` page.
- **Reusable `ErrorDisplay`** component with retry and home link.
- **Accessibility** — semantic HTML, `aria-*` attributes, keyboard navigation, focus management.
- **Smooth animations** on modals, dropdowns, and buttons.

---

## 🏗️ Architecture

The frontend follows a **hybrid rendering** pattern:

| Rendering | Used for |
|-----------|----------|
| **Server Components** | Data fetching — posts list, post detail, user profiles |
| **Client Components** | Interactivity — forms, dropdowns, modals, likes |
| **Server Actions** | Mutations — post creation |

### Project structure

```
web/
├── app/                    # Next.js App Router pages
│   ├── layout.tsx          # Root layout (Geist fonts, full-height body)
│   ├── page.tsx            # Landing page
│   ├── home/page.tsx       # Authenticated dashboard
│   ├── login/page.tsx      # Login page
│   ├── posts/              # Post CRUD pages
│   │   ├── page.tsx        # Post list
│   │   ├── [id]/page.tsx   # Post detail
│   │   └── create/         # Create post (action + page)
│   ├── [username]/page.tsx # Dynamic user profile
│   └── error/page.tsx      # Error page
├── components/
│   ├── layout/             # Navbar, ProfileDropdown, ErrorDisplay
│   ├── notification/       # NotificationDropdown, Panel, List, Item, Icon
│   ├── posts/              # CreatePostForm, LikeButton, CommentForm, etc.
│   ├── profile/            # UserProfileView, FollowListModal, ResetPasswordForm
│   └── ui/                 # InputField, Textarea
├── hooks/
│   ├── use-notifications.ts       # Notification state, WS + REST, optimistic updates
│   └── use-profile-controller.ts  # Profile state, follow, password change
├── lib/
│   ├── api.ts                     # Client-side API client (fetch wrapper)
│   ├── api-server.ts              # Server-side API client (cookie forwarding)
│   ├── notifications-api.ts       # Notification REST endpoints
│   ├── notifications-ws.ts        # WebSocket client with reconnect
│   ├── types/                     # Zod schemas + TypeScript types
│   └── utils/                     # Date formatting helpers
└── tests/                  # Vitest component and hook tests
```

---

## 🛠️ Setup

### Prerequisites

- **Node.js** 18+ and **npm**
- The Go backend running on `http://localhost:8080`

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `API_URL` | `http://localhost:8080` | Backend API base URL (used in server components) |
| `NEXT_PUBLIC_WS_URL` | `ws://localhost:8080/websocket` | WebSocket endpoint |
| `NEXT_PUBLIC_WS_DEBUG` | _(unset)_ | Set to `"true"` for verbose WS debug logging |

Set these in `.env.local`:

```
API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080/websocket
NEXT_PUBLIC_WS_DEBUG=true
```

### Install & run

```bash
npm install
npm run dev          # http://localhost:3000
npm run build        # Production build
npm run start        # Start production server
```

### Testing

```bash
npm test                    # Run Vitest
npm run test:ui             # Vitest UI mode
npm run test:coverage       # With coverage report
```

---

## 🔌 API Integration

### REST client

Two API clients handle communication with the Go backend:

- **`lib/api.ts`** — Client-side fetch wrapper with `credentials: "include"` for cookie-based auth, JSON headers, and `ApiError` class.
- **`lib/api-server.ts`** — Server-side client for Server Components / Actions. Forwards cookies from the incoming request.

### WebSocket

The notification system uses a custom WebSocket client (`lib/notifications-ws.ts`) with:

- **Auth**: JWT token sent as `?token=` query param (native `WebSocket` cannot set headers).
- **Reconnect**: Capped exponential backoff (1s → 30s, max 10 attempts).
- **Recovery**: REST fetch on reconnect to fill gaps.
- **Dedup**: By notification `id` to prevent duplicates from race conditions.

### Authentication note

> **✅ Resolved** — The backend `/websocket` handler now accepts the JWT token from three sources:
>
> 1. `Authorization: Bearer` HTTP header
> 2. `access_token` cookie (automatically forwarded by the browser)
> 3. `?token=` query parameter
>
> Native browser `WebSocket` objects use the query parameter approach.

---

## 📦 Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `next` | ^16.2.9 | Framework |
| `react` / `react-dom` | ^19.2.4 | UI library |
| `typescript` | ^5 | Language |
| `tailwindcss` | ^4 | Styling |
| `vitest` | ^4 | Testing |
| `@vitest/ui` | ^4 | Test UI |
| `@vitest/coverage-v8` | ^4 | Coverage |
| `jsdom` | ^29 | DOM environment for tests |
| `eslint` / `eslint-config-next` | — | Linting |
| `@playwright/test` | ^1.60 | E2E testing (optional) |

---

## 🧪 Test coverage

Tests are located in `web/tests/` and cover:

- **NotificationDropdown** — verify mark-as-read and mark-all-as-read button clicks.
- **ProfileDropdown** — initials rendering, dropdown toggle, logout flow (API call + redirect + localStorage clear).
- **useNotifications hook** — optimistic mark-as-read with successful API response.
