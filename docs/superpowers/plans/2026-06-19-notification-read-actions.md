# Notification Read Actions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let users mark one notification as read by clicking it and mark every notification as read with the existing panel button.

**Architecture:** Keep `useNotifications` as the only state and mutation owner. Keep the small merge/reconcile logic in that file so the diff stays compact. Keep the dropdown, panel, list, and item components presentational: they only render state and forward events. Add one hook test file, one dropdown wiring test file, and the missing Vitest setup file already referenced by `vitest.config.ts`.

**Tech Stack:** Next.js 16.2.4, React 19, TypeScript, Vitest, jsdom, existing notification API/client modules in `web/lib`.

## Global Constraints

- Clicking a notification item marks that notification as read.
- The panel button marks all notifications as read.
- The UI keeps its optimistic update behavior and rolls back on failure.
- Live websocket notifications continue to merge into the list.
- Backend API changes are out of scope.

---

## File Structure

- Modify `web/components/notification/useNotifications.ts`:
  - Keep local merge/reconcile helpers, tighten optimistic updates, and keep websocket reconciliation in one place.
- Create `web/components/notification/useNotifications.test.ts`:
  - Hook tests for optimistic success, rollback, and websocket merge behavior.
- Modify `web/components/notification/NotificationItem.tsx`:
  - Make the notification row a semantic click target so item-click behavior is testable and accessible.
- Modify `web/components/notification/NotificationPanel.tsx`:
  - Keep a single read-all button in the panel and remove the duplicate CTA.
- Modify `web/components/notification/NotificationDropdown.tsx`:
  - Keep event wiring in one place and pass the hook handlers through unchanged.
- Create `web/components/notification/NotificationDropdown.test.tsx`:
  - Component test for item click and read-all button wiring.
- Create `web/test/setup.ts`:
  - Shared Vitest setup so the current `vitest.config.ts` reference resolves and test state stays clean.

## Task 1: Tighten the Notification Hook

**Files:**
- Modify: `web/components/notification/useNotifications.ts`
- Test: `web/components/notification/useNotifications.test.ts`

**Interfaces:**
- Consumes:
  - `fetchNotifications`, `markNotificationAsRead`, `markAllNotificationsAsRead` from `web/lib/notifications-api.ts`
  - `NotificationWsClient` and `WsConnectionState` from `web/lib/notifications-ws.ts`
- Produces:
  - `useNotifications()` with the same public return shape used by the dropdown

- [ ] **Step 1: Write the failing hook test**

```tsx
import React from 'react';
import { act } from 'react-dom/test-utils';
import { createRoot } from 'react-dom/client';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useNotifications } from './useNotifications';

const fetchNotifications = vi.fn();
const markNotificationAsRead = vi.fn();
const markAllNotificationsAsRead = vi.fn();

vi.mock('@/lib/notifications-api', () => ({
  fetchNotifications,
  markNotificationAsRead,
  markAllNotificationsAsRead,
}));

vi.mock('@/lib/notifications-ws', () => ({
  NotificationWsClient: vi.fn().mockImplementation(() => ({
    connect: vi.fn(),
    disconnect: vi.fn(),
    onMessage: vi.fn(() => vi.fn()),
    onReconnect: vi.fn(),
    onStateChange: vi.fn(),
  })),
}));

function Harness({ onReady }: { onReady: (api: ReturnType<typeof useNotifications>) => void }) {
  const api = useNotifications();

  React.useEffect(() => {
    onReady(api);
  }, [api, onReady]);

  return null;
}

describe('useNotifications', () => {
  afterEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });

  it('marks one notification read optimistically and keeps the update on success', async () => {
    fetchNotifications.mockResolvedValue([
      { id: '1', title: 'One', message: 'First', type: 'info', isRead: false, createdAt: '2026-06-19T10:00:00.000Z' },
    ]);
    markNotificationAsRead.mockResolvedValue(undefined);

    let api!: ReturnType<typeof useNotifications>;
    const root = createRoot(document.createElement('div'));

    await act(async () => {
      root.render(<Harness onReady={(next) => { api = next; }} />);
    });

    await act(async () => {
      await api.markAsRead('1');
    });

    expect(markNotificationAsRead).toHaveBeenCalledWith('1');
    expect(api.notifications[0]?.isRead).toBe(true);
  });
});
```

- [ ] **Step 2: Run the hook test and confirm it fails before the hook changes**

Run: `cd web && npm test -- --run components/notification/useNotifications.test.ts`
Expected: FAIL because the current hook still needs the lazy snapshot-safe read behavior.

- [ ] **Step 3: Implement the hook changes directly in `useNotifications.ts`**

```ts
const markOneRead = (items: Notification[], id: string) =>
  items.map((item) => (item.id === id ? { ...item, isRead: true } : item));

const markAllRead = (items: Notification[]) => items.map((item) => ({ ...item, isRead: true }));

const sortByNewest = (items: Notification[]) =>
  [...items].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());

const reconcileNotifications = (live: Notification[], fresh: Notification[]) => {
  const liveMap = new Map(live.map((item) => [item.id, item]));

  for (const item of fresh) {
    if (!liveMap.has(item.id)) {
      liveMap.set(item.id, item);
    }
  }

  return sortByNewest(Array.from(liveMap.values()));
};

async function markAsRead(id: string) {
  let previous: Notification[] = [];

  setNotifications((current) => {
    previous = current;
    return markOneRead(current, id);
  });

  try {
    await markNotificationAsRead(id);
    setError(null);
  } catch (err) {
    setNotifications(previous);
    setError(err instanceof Error ? err.message : 'Failed to mark notification as read');
  }
}

async function markAllAsRead() {
  let previous: Notification[] = [];

  setNotifications((current) => {
    previous = current;
    return markAllRead(current);
  });

  try {
    await markAllNotificationsAsRead();
    setError(null);
  } catch (err) {
    setNotifications(previous);
    setError(err instanceof Error ? err.message : 'Failed to mark notifications as read');
  }
}
```

- [ ] **Step 4: Run the hook test and confirm it passes**

Run: `cd web && npm test -- --run components/notification/useNotifications.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit the hook update**

```bash
git add web/components/notification/useNotifications.ts web/components/notification/useNotifications.test.ts
git commit -m "feat: harden notification read mutations"
```

## Task 2: Simplify the UI Wiring

**Files:**
- Modify: `web/components/notification/NotificationItem.tsx`
- Modify: `web/components/notification/NotificationPanel.tsx`
- Modify: `web/components/notification/NotificationDropdown.tsx`
- Create: `web/components/notification/NotificationDropdown.test.tsx`
- Create: `web/test/setup.ts`

**Interfaces:**
- Consumes: `useNotifications()` return values from `web/components/notification/useNotifications.ts`
- Produces:
  - A semantic, clickable notification item
  - A single read-all button in the panel
  - Tests that prove the click handlers are wired through the dropdown

- [ ] **Step 1: Write the failing component test**

```tsx
import { act } from 'react-dom/test-utils';
import { createRoot } from 'react-dom/client';
import { describe, expect, it, vi } from 'vitest';
import NotificationDropdown from './NotificationDropdown';

const markAsRead = vi.fn();
const markAllAsRead = vi.fn();

vi.mock('./useNotifications', () => ({
  useNotifications: () => ({
    notifications: [
      { id: '1', title: 'New comment', message: 'Someone replied', type: 'info', isRead: false, createdAt: '2026-06-19T12:00:00.000Z' },
    ],
    isOpen: true,
    unreadCount: 1,
    connectionState: 'connected',
    toggleDropdown: vi.fn(),
    markAllAsRead,
    markAsRead,
  }),
}));

describe('NotificationDropdown', () => {
  it('forwards item clicks and read-all clicks to the hook handlers', async () => {
    const container = document.createElement('div');
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(<NotificationDropdown />);
    });

    container.querySelector('button[aria-label="Notifications"]')?.dispatchEvent(new MouseEvent('click', { bubbles: true }));
    container.querySelector('button[type="button"]')?.dispatchEvent(new MouseEvent('click', { bubbles: true }));

    expect(markAsRead).toHaveBeenCalledWith('1');
    expect(markAllAsRead).toHaveBeenCalledTimes(1);
  });
});
```

- [ ] **Step 2: Create the missing shared Vitest setup file**

```ts
import { afterEach, vi } from 'vitest';

afterEach(() => {
  vi.restoreAllMocks();
  document.body.innerHTML = '';
});
```

- [ ] **Step 3: Update `NotificationItem` to be a semantic click target**

```tsx
<button
  type="button"
  onClick={() => onClick(item.id)}
  className={`flex w-full gap-3 px-4 py-3 items-start text-left hover:bg-gray-50 dark:hover:bg-zinc-800 ${item.isRead ? 'opacity-80' : 'bg-white'}`}
>
  ...
</button>
```

- [ ] **Step 4: Keep a single read-all button in `NotificationPanel`**

```tsx
{unreadCount > 0 && (
  <div className="p-3 border-t border-gray-100 dark:border-zinc-800">
    <button onClick={onMarkAll} className="w-full rounded-md bg-indigo-600 text-white px-3 py-2 text-sm">
      Mark all as read
    </button>
  </div>
)}
```

Remove the duplicate header CTA so the panel exposes one clear action.

- [ ] **Step 5: Run the component test and confirm it passes**

Run: `cd web && npm test -- --run components/notification/NotificationDropdown.test.tsx`
Expected: PASS.

- [ ] **Step 6: Commit the UI wiring changes**

```bash
git add web/components/notification/NotificationItem.tsx web/components/notification/NotificationPanel.tsx web/components/notification/NotificationDropdown.tsx web/components/notification/NotificationDropdown.test.tsx web/test/setup.ts
git commit -m "feat: wire notification read actions in the UI"
```

## Task 3: Final Verification and Cleanup

**Files:**
- No new files expected.

**Interfaces:**
- Confirms the previous tasks produced the expected notification read behavior.

- [ ] **Step 1: Run the full web test suite**

Run: `cd web && npm test`
Expected: PASS.

- [ ] **Step 2: Run frontend linting**

Run: `cd web && npm run lint`
Expected: PASS with no new lint warnings from the notification changes.

- [ ] **Step 3: Sanity-check the notification flow manually in the browser**

Open the app and confirm all three behaviors:

```text
1. Click a single notification item and confirm only that item becomes read.
2. Click the panel's Mark all as read button and confirm every item becomes read.
3. Leave the panel open and confirm live websocket notifications still append normally.
```

- [ ] **Step 4: Commit the final verified state**

```bash
git add web/components/notification/useNotifications.ts web/components/notification/useNotifications.test.ts web/components/notification/NotificationItem.tsx web/components/notification/NotificationPanel.tsx web/components/notification/NotificationDropdown.tsx web/components/notification/NotificationDropdown.test.tsx web/test/setup.ts
git commit -m "feat: complete notification read actions integration"
```