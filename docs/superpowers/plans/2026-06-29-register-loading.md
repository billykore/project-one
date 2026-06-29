# Register Page Loading Skeleton — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create `loading.tsx` for the register page with a two-column skeleton layout mirroring the register page structure.

**Architecture:** Single new server component file (`web/app/register/loading.tsx`) that renders a two-column grid: left column is a skeleton brand panel (indigo gradient with white-translucent skeleton elements), right column is the form skeleton with 6 field placeholders, submit button, and sign-in link. Uses only the existing `Skeleton` primitive from `@/components/ui/skeleton`.

**Tech Stack:** Next.js 16 (App Router), React 19, Tailwind CSS 4, TypeScript

## Global Constraints

- Follow existing skeleton pattern: use `Skeleton` from `@/components/ui/skeleton`
- No new reusable components — bare skeleton in `loading.tsx`
- Server component only (no `"use client"` directive)
- Dark mode support via existing `Skeleton` component's `dark:bg-gray-800` for right panel; left panel uses white-translucent skeletons that work on the fixed indigo gradient
- Commit message format: `<type>(<scope>): <description>`

---

### Task 1: Create `loading.tsx` for the register page

**Files:**
- Create: `web/app/register/loading.tsx`

**Interfaces:**
- Consumes: `Skeleton` from `@/components/ui/skeleton` (existing — `interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> { className?: string }`)
- Produces: Default-exported React server component (no props)

- [ ] **Step 1: Create the loading skeleton file**

Create `web/app/register/loading.tsx`:

```tsx
import { Skeleton } from "@/components/ui/skeleton";

export default function RegisterLoading() {
  return (
    <div className="min-h-screen grid lg:grid-cols-2 font-sans">
      {/* Left: Brand Panel Skeleton */}
      <div className="relative hidden lg:flex flex-col justify-between bg-linear-to-br from-indigo-600 via-indigo-700 to-purple-800 p-12 overflow-hidden">
        {/* Decorative blurs — match real register page */}
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,rgba(255,255,255,0.1)_0%,transparent_60%)]" />
        <div className="absolute top-0 right-0 w-96 h-96 bg-linear-to-bl from-purple-500/30 to-transparent rounded-full blur-3xl" />
        <div className="absolute bottom-0 left-0 w-80 h-80 bg-linear-to-tr from-indigo-400/20 to-transparent rounded-full blur-3xl" />

        {/* Logo + brand name */}
        <div className="relative z-10">
          <div className="flex items-center gap-3">
            <Skeleton className="h-10 w-10 rounded-xl bg-white/20" />
            <Skeleton className="h-5 w-24 bg-white/20" />
          </div>
        </div>

        {/* Tagline */}
        <div className="relative z-10 space-y-3">
          <Skeleton className="h-9 w-56 bg-white/20" />
          <Skeleton className="h-9 w-48 bg-white/10" />
        </div>

        {/* Copyright */}
        <div className="relative z-10">
          <Skeleton className="h-3 w-48 bg-white/10" />
        </div>
      </div>

      {/* Right: Form Panel Skeleton */}
      <div className="flex items-center justify-center px-4 py-12 sm:px-6 lg:px-12 bg-white dark:bg-gray-950">
        <div className="w-full max-w-md space-y-8">
          {/* Mobile logo — hidden on desktop, matches real page */}
          <div className="lg:hidden flex flex-col items-center gap-2">
            <Skeleton className="h-12 w-12 rounded-xl" />
            <Skeleton className="h-5 w-24" />
          </div>

          {/* Header */}
          <div className="space-y-1">
            <Skeleton className="h-8 w-36" />
            <Skeleton className="h-4 w-64" />
          </div>

          {/* Form fields */}
          <div className="space-y-5">
            {/* First Name */}
            <Skeleton className="h-10 w-full rounded-lg" />
            {/* Last Name */}
            <Skeleton className="h-10 w-full rounded-lg" />
            {/* Username */}
            <Skeleton className="h-10 w-full rounded-lg" />
            {/* Email */}
            <Skeleton className="h-10 w-full rounded-lg" />
            {/* Password */}
            <Skeleton className="h-10 w-full rounded-lg" />
            {/* Confirm Password */}
            <Skeleton className="h-10 w-full rounded-lg" />

            {/* Submit button */}
            <Skeleton className="h-10 w-full rounded-lg" />
          </div>

          {/* Sign-in link */}
          <div className="flex justify-center">
            <Skeleton className="h-4 w-48" />
          </div>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify file compiles**

Run: `cd web && npx tsc --noEmit app/register/loading.tsx`

Expected: No errors — file compiles cleanly.

- [ ] **Step 3: Verify existing tests still pass**

Run: `cd web && npm test -- --run`

Expected: All existing tests pass (no regressions).

- [ ] **Step 4: Visual verification**

Run: `cd web && npm run dev`

Navigate to `http://localhost:3000/register` and verify:
- The loading skeleton appears briefly before the register page loads
- Two-column layout is visible on desktop (≥1024px)
- Left panel: indigo gradient with translucent skeleton placeholders
- Right panel: 6 field skeletons, button skeleton, sign-in link skeleton
- On mobile (<1024px): left panel hidden, mobile logo skeleton visible
- Dark mode: right panel skeletons darken; left panel gradient unchanged

- [ ] **Step 5: Commit**

```bash
git add web/app/register/loading.tsx
git commit -m "feat(register): add loading skeleton for registration page"
```
