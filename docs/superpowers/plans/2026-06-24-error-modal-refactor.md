# Error Modal Refactor — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the dedicated `/error` route page with a client-side modal overlay triggered via React Context.

**Architecture:** An `ErrorModalProvider` wraps the root layout, providing `showError(message, onRetry?)` and `closeError()` via context. The `ErrorModal` component renders as a fixed overlay with backdrop, two buttons ("Try again" and "Go back home"), and full accessibility. The existing `ErrorDisplay` component is untouched.

**Tech Stack:** Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS v4, Vitest + jsdom

## Global Constraints

- Next.js 16 — check `web/node_modules/next/dist/docs/` for any breaking changes before writing code
- All new components use `"use client"` directive
- Follow existing modal patterns (`FollowListModal`) for backdrop, Escape key, `role="dialog"`
- Existing `ErrorDisplay` component (`web/components/layout/error.tsx`) must NOT be modified
- Tests follow existing pattern: `createRoot` + `act`, no Testing Library render

---

## Task Summary

| # | Task | Files |
|---|------|-------|
| 1 | Create `useErrorModal` hook + context | `web/hooks/use-error-modal.ts` |
| 2 | Test `useErrorModal` hook | `web/tests/hooks/use-error-modal.test.tsx` |
| 3 | Create `ErrorModal` component | `web/components/layout/error-modal.tsx` |
| 4 | Test `ErrorModal` component | `web/tests/components/error-modal.test.tsx` |
| 5 | Wire into root layout | `web/app/layout.tsx` |
| 6 | Update login page caller | `web/app/(auth)/login/page.tsx` |
| 7 | Delete old error route | `web/app/error/page.tsx` |
| 8 | Final verification | Run full test suite, manual check |

---

### Task 1: Create `useErrorModal` Hook + Context

**Files:**
- Create: `web/hooks/use-error-modal.ts`

**Interfaces:**
- Produces: `ErrorModalProvider` (component), `useErrorModal` (hook)
- `ErrorModalProvider` wraps children and provides context
- `useErrorModal()` returns `{ showError: (message: string, onRetry?: () => void) => void; closeError: () => void }`

- [ ] **Step 1: Create `web/hooks/use-error-modal.ts`**

```ts
"use client";

import React, { createContext, useContext, useState, useCallback } from "react";

interface ErrorModalState {
  open: boolean;
  message: string;
  onRetry?: () => void;
}

interface ErrorModalContextValue {
  showError: (message: string, onRetry?: () => void) => void;
  closeError: () => void;
  state: ErrorModalState;
}

const ErrorModalContext = createContext<ErrorModalContextValue | null>(null);

export function ErrorModalProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<ErrorModalState>({
    open: false,
    message: "",
    onRetry: undefined,
  });

  const showError = useCallback((message: string, onRetry?: () => void) => {
    if (!message) return;
    setState({ open: true, message, onRetry });
  }, []);

  const closeError = useCallback(() => {
    setState({ open: false, message: "", onRetry: undefined });
  }, []);

  return (
    <ErrorModalContext.Provider value={{ showError, closeError, state }}>
      {children}
    </ErrorModalContext.Provider>
  );
}

export function useErrorModal(): ErrorModalContextValue {
  const ctx = useContext(ErrorModalContext);
  if (!ctx) {
    throw new Error("useErrorModal must be used within an ErrorModalProvider");
  }
  return ctx;
}
```

- [ ] **Step 2: Verify file compiles**

Run: `cd web && npx tsc --noEmit hooks/use-error-modal.ts 2>&1 | head -20`
Expected: No errors (may show module resolution warnings which are OK in isolation)

- [ ] **Step 3: Commit**

```bash
git add web/hooks/use-error-modal.ts
git commit -m "feat: add useErrorModal hook and context provider"
```

---

### Task 2: Test `useErrorModal` Hook

**Files:**
- Create: `web/tests/hooks/use-error-modal.test.tsx`

**Interfaces:**
- Consumes: `ErrorModalProvider`, `useErrorModal` from `web/hooks/use-error-modal.ts`

- [ ] **Step 1: Create `web/tests/hooks/use-error-modal.test.tsx`**

```ts
import React from "react";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { describe, expect, it, beforeEach } from "vitest";
import { ErrorModalProvider, useErrorModal } from "@/hooks/use-error-modal";

function TestConsumer({
  onRender,
}: {
  onRender: (ctx: ReturnType<typeof useErrorModal>) => void;
}) {
  const ctx = useErrorModal();
  onRender(ctx);
  return null;
}

describe("useErrorModal", () => {
  let container: HTMLDivElement;
  let captured: ReturnType<typeof useErrorModal> | null = null;
  const capture = (ctx: ReturnType<typeof useErrorModal>) => {
    captured = ctx;
  };

  beforeEach(() => {
    captured = null;
    container = document.createElement("div");
    document.body.appendChild(container);
  });

  it("throws when used outside provider", () => {
    expect(() => {
      createRoot(container).render(<TestConsumer onRender={capture} />);
    }).toThrow("useErrorModal must be used within an ErrorModalProvider");
  });

  it("sets open=true and message when showError is called", async () => {
    await act(async () => {
      createRoot(container).render(
        <ErrorModalProvider>
          <TestConsumer onRender={capture} />
        </ErrorModalProvider>
      );
    });

    await act(async () => {
      captured!.showError("Test error message");
    });

    expect(captured!.state.open).toBe(true);
    expect(captured!.state.message).toBe("Test error message");
  });

  it("does not open when showError is called with empty message", async () => {
    await act(async () => {
      createRoot(container).render(
        <ErrorModalProvider>
          <TestConsumer onRender={capture} />
        </ErrorModalProvider>
      );
    });

    await act(async () => {
      captured!.showError("");
    });

    expect(captured!.state.open).toBe(false);
  });

  it("stores onRetry callback when provided", async () => {
    await act(async () => {
      createRoot(container).render(
        <ErrorModalProvider>
          <TestConsumer onRender={capture} />
        </ErrorModalProvider>
      );
    });

    const retryFn = () => {};
    await act(async () => {
      captured!.showError("msg", retryFn);
    });

    expect(captured!.state.onRetry).toBe(retryFn);
  });

  it("closeError resets state to closed", async () => {
    await act(async () => {
      createRoot(container).render(
        <ErrorModalProvider>
          <TestConsumer onRender={capture} />
        </ErrorModalProvider>
      );
    });

    await act(async () => {
      captured!.showError("msg", () => {});
    });
    expect(captured!.state.open).toBe(true);

    await act(async () => {
      captured!.closeError();
    });

    expect(captured!.state.open).toBe(false);
    expect(captured!.state.message).toBe("");
    expect(captured!.state.onRetry).toBeUndefined();
  });

  it("replaces existing error when showError is called while open", async () => {
    await act(async () => {
      createRoot(container).render(
        <ErrorModalProvider>
          <TestConsumer onRender={capture} />
        </ErrorModalProvider>
      );
    });

    const firstRetry = () => {};
    const secondRetry = () => {};

    await act(async () => {
      captured!.showError("first", firstRetry);
    });
    expect(captured!.state.message).toBe("first");

    await act(async () => {
      captured!.showError("second", secondRetry);
    });
    expect(captured!.state.message).toBe("second");
    expect(captured!.state.onRetry).toBe(secondRetry);
    expect(captured!.state.open).toBe(true);
  });
});
```

- [ ] **Step 2: Run tests to verify they fail (hook file exists but tests are new)**

Run: `cd web && npx vitest run tests/hooks/use-error-modal.test.tsx 2>&1`
Expected: All 6 tests pass (hook already created in Task 1)

- [ ] **Step 3: Commit**

```bash
git add web/tests/hooks/use-error-modal.test.tsx
git commit -m "test: add useErrorModal hook tests"
```

---

### Task 3: Create `ErrorModal` Component

**Files:**
- Create: `web/components/layout/error-modal.tsx`

**Interfaces:**
- Consumes: `useErrorModal` from `web/hooks/use-error-modal.ts`
- Consumes: Next.js `Link` from `next/link`
- Produces: `ErrorModal` component (no props, reads from context)

- [ ] **Step 1: Create `web/components/layout/error-modal.tsx`**

```tsx
"use client";

import React, { useEffect, useRef } from "react";
import Link from "next/link";
import { useErrorModal } from "@/hooks/use-error-modal";

export default function ErrorModal() {
  const { state, closeError } = useErrorModal();
  const { open, message, onRetry } = state;
  const primaryBtnRef = useRef<HTMLButtonElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && open) {
        closeError();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [open, closeError]);

  useEffect(() => {
    if (open) {
      previousFocusRef.current = document.activeElement as HTMLElement;
      primaryBtnRef.current?.focus();
    } else if (previousFocusRef.current) {
      previousFocusRef.current.focus();
    }
  }, [open]);

  if (!open) return null;

  const handleRetry = () => {
    if (onRetry) {
      onRetry();
    } else {
      window.location.reload();
    }
    closeError();
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm transition-opacity"
      onClick={closeError}
      role="dialog"
      aria-modal="true"
      aria-labelledby="error-modal-title"
      aria-describedby="error-modal-message"
    >
      <div
        className="w-full max-w-md rounded-xl bg-white p-8 shadow-xl dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 transform transition-all"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-6 flex justify-center">
          <svg
            className="h-12 w-12 text-red-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
        </div>

        <h2
          id="error-modal-title"
          className="text-center text-2xl font-bold text-zinc-900 dark:text-zinc-50 mb-2"
        >
          Something went wrong
        </h2>

        <p
          id="error-modal-message"
          className="text-center text-zinc-600 dark:text-zinc-400 mb-8"
        >
          {message}
        </p>

        <div className="flex flex-col gap-3">
          <button
            ref={primaryBtnRef}
            onClick={handleRetry}
            className="flex h-11 items-center justify-center rounded-full bg-zinc-900 px-8 text-sm font-medium text-white transition-colors hover:bg-zinc-700 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200 cursor-pointer"
          >
            Try again
          </button>
          <Link
            href="/"
            onClick={closeError}
            className="flex h-11 items-center justify-center rounded-full border border-zinc-200 px-8 text-sm font-medium text-zinc-900 transition-colors hover:bg-zinc-50 dark:border-zinc-800 dark:text-zinc-50 dark:hover:bg-zinc-900"
          >
            Go back home
          </Link>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify file compiles**

Run: `cd web && npx tsc --noEmit 2>&1 | head -20`
Expected: No new TypeScript errors

- [ ] **Step 3: Commit**

```bash
git add web/components/layout/error-modal.tsx
git commit -m "feat: add ErrorModal component with retry and home buttons"
```

---

### Task 4: Test `ErrorModal` Component

**Files:**
- Create: `web/tests/components/error-modal.test.tsx`

**Interfaces:**
- Consumes: `ErrorModal` from `web/components/layout/error-modal.tsx`
- Consumes: `ErrorModalProvider` from `web/hooks/use-error-modal.ts`

- [ ] **Step 1: Create `web/tests/components/error-modal.test.tsx`**

```ts
import React from "react";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { ErrorModalProvider, useErrorModal } from "@/hooks/use-error-modal";
import ErrorModal from "@/components/layout/error-modal";

// Mock next/link
vi.mock("next/link", () => ({
  default: ({
    children,
    href,
    onClick,
    className,
  }: {
    children: React.ReactNode;
    href: string;
    onClick?: () => void;
    className?: string;
  }) => (
    <a href={href} onClick={onClick} className={className}>
      {children}
    </a>
  ),
}));

// Mock window.location.reload
const reloadMock = vi.fn();
Object.defineProperty(window, "location", {
  value: { reload: reloadMock },
  writable: true,
});

function TestHarness({ children }: { children: React.ReactNode }) {
  return <ErrorModalProvider>{children}</ErrorModalProvider>;
}

function TriggerButton() {
  const { showError } = useErrorModal();
  return (
    <button onClick={() => showError("Test error", undefined)}>
      Trigger Error
    </button>
  );
}

describe("ErrorModal", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("does not render when no error is set", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
        </TestHarness>
      );
    });

    expect(container.querySelector('[role="dialog"]')).toBeNull();
  });

  it("renders modal with message when showError is called", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
          <TriggerButton />
        </TestHarness>
      );
    });

    await act(async () => {
      (container.querySelector("button") as HTMLButtonElement).click();
    });

    const dialog = container.querySelector('[role="dialog"]');
    expect(dialog).not.toBeNull();
    expect(dialog?.getAttribute("aria-modal")).toBe("true");
    expect(container.textContent).toContain("Test error");
    expect(container.textContent).toContain("Something went wrong");
    expect(container.textContent).toContain("Try again");
    expect(container.textContent).toContain("Go back home");
  });

  it("closes modal when backdrop is clicked", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
          <TriggerButton />
        </TestHarness>
      );
    });

    await act(async () => {
      (container.querySelector("button") as HTMLButtonElement).click();
    });

    const backdrop = container.querySelector('[role="dialog"]') as HTMLElement;
    await act(async () => {
      backdrop.click();
    });

    expect(container.querySelector('[role="dialog"]')).toBeNull();
  });

  it("does not close when dialog inner is clicked", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
          <TriggerButton />
        </TestHarness>
      );
    });

    await act(async () => {
      (container.querySelector("button") as HTMLButtonElement).click();
    });

    // Click the inner dialog box (not the backdrop)
    const inner = container.querySelector(".max-w-md") as HTMLElement;
    await act(async () => {
      inner.click();
    });

    expect(container.querySelector('[role="dialog"]')).not.toBeNull();
  });

  it("closes modal when Escape key is pressed", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
          <TriggerButton />
        </TestHarness>
      );
    });

    await act(async () => {
      (container.querySelector("button") as HTMLButtonElement).click();
    });

    await act(async () => {
      window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    });

    expect(container.querySelector('[role="dialog"]')).toBeNull();
  });

  it("calls onRetry callback when Try again is clicked", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    const retryFn = vi.fn();

    function TriggerWithRetry() {
      const { showError } = useErrorModal();
      return (
        <button onClick={() => showError("msg", retryFn)}>
          Trigger
        </button>
      );
    }

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
          <TriggerWithRetry />
        </TestHarness>
      );
    });

    await act(async () => {
      (container.querySelector("button") as HTMLButtonElement).click();
    });

    const tryAgainBtn = Array.from(container.querySelectorAll("button")).find(
      (b) => b.textContent === "Try again"
    ) as HTMLButtonElement;

    await act(async () => {
      tryAgainBtn.click();
    });

    expect(retryFn).toHaveBeenCalledTimes(1);
    expect(container.querySelector('[role="dialog"]')).toBeNull();
  });

  it("reloads page when Try again is clicked and no onRetry provided", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
          <TriggerButton />
        </TestHarness>
      );
    });

    await act(async () => {
      (container.querySelector("button") as HTMLButtonElement).click();
    });

    const tryAgainBtn = Array.from(container.querySelectorAll("button")).find(
      (b) => b.textContent === "Try again"
    ) as HTMLButtonElement;

    await act(async () => {
      tryAgainBtn.click();
    });

    expect(reloadMock).toHaveBeenCalledTimes(1);
  });

  it("navigates home and closes when Go back home is clicked", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
          <TriggerButton />
        </TestHarness>
      );
    });

    await act(async () => {
      (container.querySelector("button") as HTMLButtonElement).click();
    });

    const homeLink = container.querySelector('a[href="/"]') as HTMLAnchorElement;

    await act(async () => {
      homeLink.click();
    });

    expect(container.querySelector('[role="dialog"]')).toBeNull();
  });
});
```

- [ ] **Step 2: Run tests**

Run: `cd web && npx vitest run tests/components/error-modal.test.tsx 2>&1`
Expected: All 8 tests pass

- [ ] **Step 3: Commit**

```bash
git add web/tests/components/error-modal.test.tsx
git commit -m "test: add ErrorModal component tests"
```

---

### Task 5: Wire `ErrorModalProvider` + `ErrorModal` into Root Layout

**Files:**
- Modify: `web/app/layout.tsx`

**Interfaces:**
- Consumes: `ErrorModalProvider` from `web/hooks/use-error-modal.ts`
- Consumes: `ErrorModal` from `web/components/layout/error-modal.tsx`

- [ ] **Step 1: Modify `web/app/layout.tsx`**

Add the import and wrap the body children. Change the body content from:

```tsx
      <body className="min-h-full flex flex-col">{children}</body>
```

To:

```tsx
      <body className="min-h-full flex flex-col">
        <ErrorModalProvider>
          {children}
          <ErrorModal />
        </ErrorModalProvider>
      </body>
```

And add the imports at the top:

```tsx
import { ErrorModalProvider } from "@/hooks/use-error-modal";
import ErrorModal from "@/components/layout/error-modal";
```

- [ ] **Step 2: Verify compilation**

Run: `cd web && npx tsc --noEmit 2>&1 | head -20`
Expected: No TypeScript errors

- [ ] **Step 3: Commit**

```bash
git add web/app/layout.tsx
git commit -m "feat: wire ErrorModalProvider and ErrorModal into root layout"
```

---

### Task 6: Update Login Page to Use `showError`

**Files:**
- Modify: `web/app/(auth)/login/page.tsx` (lines ~1-8, ~92)

**Interfaces:**
- Consumes: `useErrorModal` from `web/hooks/use-error-modal.ts`

- [ ] **Step 1: Replace `router.push` with `showError`**

In `web/app/(auth)/login/page.tsx`, change the error handler. The current code at ~line 92:

```tsx
      const errorMessage = err instanceof Error ? err.message : "An error occurred. Please try again later.";
      router.push(`/error?message=${encodeURIComponent(errorMessage)}`);
```

Replace with:

```tsx
      const errorMessage = err instanceof Error ? err.message : "An error occurred. Please try again later.";
      showError(errorMessage);
```

- [ ] **Step 2: Add the `useErrorModal` import and hook call**

Add at the top of the file with other imports:

```tsx
import { useErrorModal } from "@/hooks/use-error-modal";
```

Inside the `LoginPage` component, after `const router = useRouter();` add:

```tsx
  const { showError } = useErrorModal();
```

- [ ] **Step 3: Verify compilation**

Run: `cd web && npx tsc --noEmit 2>&1 | head -20`
Expected: No TypeScript errors

- [ ] **Step 4: Commit**

```bash
git add web/app/\(auth\)/login/page.tsx
git commit -m "refactor: replace error page navigation with showError modal in login"
```

---

### Task 7: Delete Old Error Route Page

**Files:**
- Delete: `web/app/error/page.tsx`

- [ ] **Step 1: Delete the file**

```bash
rm web/app/error/page.tsx
```

- [ ] **Step 2: Remove empty directory if applicable**

```bash
rmdir web/app/error 2>/dev/null; true
```

- [ ] **Step 3: Verify no remaining references**

Run: `cd web && grep -r "error/page" --include="*.ts" --include="*.tsx" . 2>/dev/null`
Expected: No output (no remaining imports or references)

- [ ] **Step 4: Commit**

```bash
git add web/app/error/
git commit -m "refactor: delete old /error route page"
```

---

### Task 8: Final Verification

- [ ] **Step 1: Run full test suite**

```bash
cd web && npx vitest run 2>&1
```
Expected: All tests pass

- [ ] **Step 2: Type check entire project**

```bash
cd web && npx tsc --noEmit 2>&1
```
Expected: No errors

- [ ] **Step 3: Build check**

```bash
cd web && npx next build 2>&1 | tail -20
```
Expected: Successful build, no errors related to the error modal changes

- [ ] **Step 4: Commit (if any cleanup needed)**

```bash
git add -A
git commit -m "chore: final verification after error modal refactor"
```
