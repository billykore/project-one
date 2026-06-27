# Login Page Loading Skeleton — Design Spec

**Date:** 2026-06-27
**Status:** Approved

## Goal

Create a `loading.tsx` for the login page (`/login`) that displays skeleton loading states while the page content loads, consistent with existing skeleton patterns in the project.

## Approach

**Approach 1 — Bare `Skeleton` in `loading.tsx` (selected)**

Write `loading.tsx` directly with raw `<Skeleton>` elements in a centered layout. No new reusable component — the simplest path.

*Rationale:* The login page has a unique two-panel layout (brand + form) that differs from other pages. A standalone skeleton avoids coupling to `PageSkeletonLayout` (which adds a navbar the login page doesn't have) and keeps the change minimal.

## Architecture

### New File

**`web/app/login/loading.tsx`** — Server component (no `"use client"` needed — no interactivity).

### Dependencies

- Imports `Skeleton` from `@/components/ui/skeleton`
- No changes to any existing file

## Layout

Centered card matching the login form panel's visual proportions:

```
┌──────────────────────────────┐
│                              │
│       [logo icon]     <- lg:hidden (mobile only)
│       [project name]  <- lg:hidden (mobile only)
│                              │
│       [ Sign in     ]  h-8   │
│       [ subtitle    ]  h-4   │
│                              │
│       [ email field ]  h-10  │
│                              │
│       [ password    ]  h-10  │
│                              │
│       [link]  [link]    h-4  │
│                              │
│       [ submit btn  ]  h-10  │
│                              │
└──────────────────────────────┘
```

Container: `min-h-screen flex items-center justify-center`, `max-w-md`, with spacing matching the real form (`space-y-5`).

## Skeleton Elements

| Element | Skeleton classes | Real counterpart |
|---------|-----------------|------------------|
| Logo icon | `h-12 w-12 rounded-xl` | Gradient logo box |
| Project name | `h-5 w-24` | "Project1" text |
| Form heading | `h-8 w-24` | "Sign in" h2 |
| Subtitle | `h-4 w-64` | "Enter your credentials…" |
| Email field | `h-10 w-full rounded-lg` | `<InputField>` |
| Password field | `h-10 w-full rounded-lg` | `<InputField>` |
| Links row | `h-4 w-32` + `h-4 w-28` | Forgot password / Create account |
| Submit button | `h-10 w-full rounded-lg` | "Sign in" button |

## Edge Cases

- **Dark mode:** `Skeleton` already handles `dark:bg-gray-800` — no extra work needed
- **Mobile:** Logo skeleton uses `lg:hidden` to only show on small screens, mirroring the real login page's mobile-only logo
- **Wrapping:** The `loading.tsx` is a server component — Next.js automatically shows it during navigation and initial page load

## Non-Goals

- No new reusable skeleton component in `page-skeleton.tsx`
- No changes to the login page itself
- No changes to the `Skeleton` primitive component
