# Error Modal Refactor — Design Spec

**Date**: 2026-06-24
**Status**: Draft

## Overview

Refactor the dedicated `/error` route page into a client-side modal overlay. The error message is passed directly (not via query parameter), and the modal shows two always-visible buttons: "Try again" (calls an optional `onRetry` callback, defaults to page reload) and "Go back home" (navigates to `/`).

## Architecture

### Files Changed

| File | Action | Purpose |
|------|--------|---------|
| `web/hooks/use-error-modal.ts` | ✨ New | Context, provider, and `useErrorModal` hook |
| `web/components/layout/error-modal.tsx` | ✨ New | Modal UI component |
| `web/app/layout.tsx` | ✏️ Modify | Wrap children in `ErrorModalProvider`, render `<ErrorModal />` |
| `web/app/(auth)/login/page.tsx` | ✏️ Modify | `router.push('/error?...')` → `showError(message, onRetry)` |
| `web/app/error/page.tsx` | ❌ Delete | No longer needed |
| `web/components/layout/error.tsx` | 🔒 Unchanged | Stays for Next.js error boundaries |

### Component Tree

```
RootLayout
└── ErrorModalProvider           ← context: { open, message, onRetry }
    ├── {children}               ← app pages (login, posts, etc.)
    └── ErrorModal               ← conditionally renders overlay
```

## Data Flow

### Context API

```ts
interface ErrorModalState {
  open: boolean;
  message: string;
  onRetry?: () => void;
}

useErrorModal() → {
  showError(message: string, onRetry?: () => void): void;
  closeError(): void;
}
```

### Sequence

1. **Caller** calls `showError("message", optionalRetryCallback)`
2. **Context** sets `{ open: true, message, onRetry }`, modal renders
3. **User** clicks "Try again" → `onRetry()` (or `window.location.reload()` if not provided) → `closeError()`
4. **User** clicks "Go back home" → navigate to `/` → `closeError()`
5. **User** presses Escape or clicks backdrop → `closeError()`

### Behaviors

- **Empty/blank message**: Modal does not open.
- **`onRetry` default**: `window.location.reload()`.
- **Re-opening**: Calling `showError` while open replaces the existing message and callback (no stacking).
- **Clearing on close**: `open` → `false`, `message` and `onRetry` cleared.

## Modal UI

### Layout

```
┌──────────────────────────────────────────┐
│         (dark semi-transparent backdrop)   │
│                                            │
│   ┌────────────────────────────────┐       │
│   │  ⚠️  warning icon (red-500)    │       │
│   │                                │       │
│   │  Something went wrong          │       │
│   │  <error message>               │       │
│   │                                │       │
│   │  [ Try again ]    (primary)    │       │
│   │  [ Go back home ] (secondary)  │       │
│   └────────────────────────────────┘       │
└──────────────────────────────────────────┘
```

### Styling

Follows existing design system (Tailwind CSS v4, light/dark themes):

- **Backdrop**: `fixed inset-0 z-50 bg-black/40 backdrop-blur-sm`
- **Dialog**: `max-w-md bg-white dark:bg-zinc-900 rounded-xl shadow-xl border border-zinc-200 dark:border-zinc-800 p-8`
- **"Try again"**: Filled rounded-full button (zinc-900 / dark:zinc-50)
- **"Go back home"**: Outline rounded-full button (border-zinc-200 / dark:border-zinc-800)
- **Warning icon**: Reuses the existing SVG triangle icon from `ErrorDisplay`

### Accessibility

- `role="dialog"`, `aria-modal="true"` on backdrop
- `aria-labelledby` → heading, `aria-describedby` → message text
- Escape key closes modal
- Backdrop click closes modal; inner dialog `stopPropagation()`
- Focus trap: primary button focused on open; focus returns to trigger on close

## Caller Migration

### Before (login page)

```ts
router.push(`/error?message=${encodeURIComponent(errorMessage)}`);
```

### After (login page)

```ts
const { showError } = useErrorModal();
// ...
showError(errorMessage, () => window.location.reload());
```

## Edge Cases

1. **Modal opens with no `onRetry`**: "Try again" still visible, calls `window.location.reload()`.
2. **Rapid double-open**: Second call replaces first — no stacking.
3. **Navigation while modal is open**: Closing on navigation is handled naturally since the provider unmounts on route change (Next.js layout persists, so the modal stays — this is acceptable; the user dismissing it is the expected path).
4. **`ErrorDisplay` consumers unaffected**: The error boundaries in `posts/error.tsx` and `posts/[id]/error.tsx` continue using the existing `ErrorDisplay` card component. They are separate concerns.
5. **Server components**: The hook and modal are `"use client"`. Server components cannot call `showError` directly, but none currently do.

## Testing

- **Unit**: `useErrorModal` hook — verify `showError` sets state, `closeError` clears it, empty message doesn't open.
- **Component**: `ErrorModal` — renders when `open=true`, calls `onRetry` on button click, calls `closeError` on Escape/backdrop.
- **Integration**: Login page — on auth failure, modal opens with the error message instead of navigating away.
