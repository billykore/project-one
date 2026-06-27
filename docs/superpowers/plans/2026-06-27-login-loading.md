# Login Page Loading Skeleton — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create `loading.tsx` for the login page with skeleton loading states.

**Architecture:** Single new server component file (`web/app/login/loading.tsx`) that renders skeleton elements using the existing `Skeleton` primitive from `@/components/ui/skeleton`. No modifications to existing files.

**Tech Stack:** Next.js 16 (App Router), React 19, Tailwind CSS 4, TypeScript

## Global Constraints

- Follow existing skeleton pattern: use `Skeleton` from `@/components/ui/skeleton`
- No new reusable components — bare skeleton in `loading.tsx`
- Server component only (no `"use client"` directive)
- Dark mode support via existing `Skeleton` component's `dark:bg-gray-800`
- Commit message format: `<type>(<scope>): <description>`

---

### Task 1: Create `loading.tsx` for the login page

**Files:**
- Create: `web/app/login/loading.tsx`

**Interfaces:**
- Consumes: `Skeleton` from `@/components/ui/skeleton` (existing — `interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> { className?: string }`)
- Produces: Default-exported React server component (no props)

- [ ] **Step 1: Create the loading skeleton file**

Create `web/app/login/loading.tsx`:

```tsx
import { Skeleton } from "@/components/ui/skeleton";

export default function LoginLoading() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-white px-4 py-12 dark:bg-gray-950 sm:px-6 lg:px-12">
      <div className="w-full max-w-md space-y-5">
        {/* Mobile logo (visible only on small screens, mirrors real login page) */}
        <div className="flex flex-col items-center gap-2 lg:hidden">
          <Skeleton className="h-12 w-12 rounded-xl" />
          <Skeleton className="h-5 w-24" />
        </div>

        {/* Form heading */}
        <div className="space-y-1">
          <Skeleton className="h-8 w-24" />
          <Skeleton className="h-4 w-64" />
        </div>

        {/* Email field */}
        <Skeleton className="h-10 w-full rounded-lg" />

        {/* Password field */}
        <Skeleton className="h-10 w-full rounded-lg" />

        {/* Links row */}
        <div className="flex items-center justify-between pt-1">
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-4 w-28" />
        </div>

        {/* Submit button */}
        <Skeleton className="h-10 w-full rounded-lg" />
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify file compiles**

Run: `cd web && npx tsc --noEmit app/login/loading.tsx`
Expected: No errors — file compiles cleanly.

- [ ] **Step 3: Commit**

```bash
git add web/app/login/loading.tsx
git commit -m "feat(login): add loading skeleton for login page"
```
