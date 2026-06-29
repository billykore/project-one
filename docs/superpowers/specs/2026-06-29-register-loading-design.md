# Register Page Loading Skeleton — Design Spec

**Date:** 2026-06-29
**Status:** Approved

## Goal

Create a `loading.tsx` for the register page (`/register`) that displays skeleton loading states while the page content loads, consistent with existing skeleton patterns in the project.

## Approach

**Approach 1 — Two-column mirror with `Skeleton` (selected)**

Write `loading.tsx` that mirrors the register page's two-column grid layout. Left column is a skeleton brand panel (indigo gradient with placeholder elements), right column is the form skeleton with 6 input fields, submit button, and sign-in link. Uses only the existing `Skeleton` component — no new reusable components.

*Rationale:* The register page has a unique two-panel layout (brand panel + form) that `PageSkeletonLayout` doesn't model. The login loading skeleton also uses a standalone `Skeleton`-based approach for the same reason. Mirroring the full two-column layout gives users an accurate preview of the page structure while it loads.

**Approach 2 — Centered form only (rejected)**

Simpler single-column centered skeleton like `login/loading.tsx`. *Rejected:* User preferred the two-column mirror for visual fidelity.

## Architecture

### New File

**`web/app/register/loading.tsx`** — Server component (no `"use client"` needed — no interactivity).

### Dependencies

- Imports `Skeleton` from `@/components/ui/skeleton`
- No changes to any existing file

## Layout

Two-column grid matching the register page's `lg:grid-cols-2`:

```
┌──────────────────────┬──────────────────────────────┐
│  Brand Panel (left)  │  Form Panel (right)          │
│  hidden on mobile    │                              │
│                      │  [logo icon]    lg:hidden    │
│  [logo icon]         │  [project name] lg:hidden    │
│  [project name]      │                              │
│                      │  [ Create account  ]  h-8    │
│  [tagline line 1]    │  [ subtitle        ]  h-4    │
│  [tagline line 2]    │                              │
│                      │  [ first name field ]  h-10  │
│  [copyright text]    │  [ last name field  ]  h-10  │
│                      │  [ username field   ]  h-10  │
│                      │  [ email field      ]  h-10  │
│                      │  [ password field   ]  h-10  │
│                      │  [ confirm pwd field]  h-10  │
│                      │                              │
│                      │  [ submit button    ]  h-10  │
│                      │  [ sign-in link     ]  h-4   │
└──────────────────────┴──────────────────────────────┘
```

Left panel: Same indigo-to-purple gradient background as the real page, with absolute-positioned decorative blurs.

Right panel: `flex items-center justify-center`, `max-w-md`, with spacing matching the real form (`space-y-5` for fields, `space-y-8` for the form wrapper).

## Skeleton Elements

### Left (Brand Panel)

| Element | Skeleton classes | Real counterpart |
|---------|-----------------|------------------|
| Logo icon | `h-10 w-10 rounded-xl bg-white/20` | White/translucent logo box |
| Project name | `h-5 w-24 bg-white/20` | "Project1" text |
| Tagline line 1 | `h-9 w-56 bg-white/20` | "Join the community." |
| Tagline line 2 | `h-9 w-48 bg-white/10` | "Create your account." |
| Copyright | `h-3 w-48 bg-white/10` | Footer text |

### Right (Form Panel)

| Element | Skeleton classes | Real counterpart |
|---------|-----------------|------------------|
| Mobile logo icon | `h-12 w-12 rounded-xl` (default Skeleton) | Gradient logo box |
| Mobile project name | `h-5 w-24` | "Project1" text |
| Heading | `h-8 w-36` | "Create account" h2 |
| Subtitle | `h-4 w-64` | Description text |
| First Name field | `h-10 w-full rounded-lg` | InputField |
| Last Name field | `h-10 w-full rounded-lg` | InputField |
| Username field | `h-10 w-full rounded-lg` | InputField |
| Email field | `h-10 w-full rounded-lg` | InputField |
| Password field | `h-10 w-full rounded-lg` | InputField |
| Confirm Password field | `h-10 w-full rounded-lg` | InputField |
| Submit button | `h-10 w-full rounded-lg` | Button |
| Sign-in link | `h-4 w-48` | Link text |

## Skeleton Color Notes

- Left panel skeletons use `bg-white/20` and `bg-white/10` to blend with the dark gradient background (since `Skeleton`'s default gray won't be visible on dark indigo).
- Right panel skeletons use the default `Skeleton` component (gray on light bg, dark-gray on dark bg).
- Elements wrapped in `<div aria-hidden="true">` via the Skeleton component itself.

## Edge Cases

- **Dark mode:** The `Skeleton` component already supports dark mode via `dark:bg-gray-800`. Left panel gradient is the same in both themes (no dark variant needed — consistent with the real page).
- **Mobile:** Left brand panel hidden (`hidden lg:flex`), matching the real page.
- **Long loads:** The `animate-pulse` from the Skeleton component provides continuous visual feedback.
