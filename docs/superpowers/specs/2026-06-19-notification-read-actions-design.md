# Notification Read Actions UI Design

## Objective

Integrate the notification UI with the existing notification APIs so users can mark a single notification as read by clicking it, and mark all notifications as read using the existing panel button.

This design keeps the current notification stack intact and focuses on wiring, state updates, and failure handling. The backend already exposes the required endpoints:

- `GET /notifications`
- `PUT /notifications/{id}/read`
- `PUT /notifications/read-all`

## Scope

### In scope

- Clicking a notification item marks that notification as read.
- The panel button marks all notifications as read.
- The UI keeps its optimistic update behavior and rolls back on failure.
- Live websocket notifications continue to merge into the list.

### Out of scope

- Backend API changes.
- New notification types or schemas.
- New notification entry points outside the existing panel.

## Current State

The frontend already has the core pieces in place:

- `useNotifications` owns notification state, unread count, loading, error, and websocket lifecycle.
- `NotificationDropdown` renders the icon and panel.
- `NotificationList` renders each notification item.
- `NotificationItem` already calls the item click handler.
- `notifications-api.ts` already provides fetch, mark-one-read, and mark-all-read functions.

Because the plumbing exists, the work is primarily about making the behavior explicit and ensuring the UI contract stays consistent.

## Design

### State ownership

`useNotifications` remains the single source of truth for notification state and mutation behavior.

It will continue to own:

- The list of notifications.
- The unread count.
- The open/closed state of the dropdown.
- Websocket connection state.
- Error and loading state.
- Optimistic updates for read actions.

### Single-item read flow

When the user clicks a notification item:

1. The item click handler receives the notification ID.
2. `useNotifications` immediately updates that item to `isRead: true` in local state.
3. The hook calls `PUT /notifications/{id}/read` through `markNotificationAsRead`.
4. On success, the optimistic state remains.
5. On failure, the hook restores the previous list and exposes the error message.

This keeps the interaction fast while preserving correctness if the request fails.

### Read-all flow

When the user clicks the existing `Mark all as read` button:

1. `useNotifications` marks every notification as read locally.
2. The hook calls `PUT /notifications/read-all` through `markAllNotificationsAsRead`.
3. On success, the optimistic state remains.
4. On failure, the hook restores the previous list and exposes the error message.

The button stays in the notification panel header/footer so the interaction remains discoverable and consistent with the existing layout.

### Realtime merging

Incoming websocket notifications continue to flow through the existing `NotificationWsClient`.

The merge behavior should preserve these rules:

- New notifications are inserted once, not duplicated.
- Fresh websocket notifications can appear while the panel is open or closed.
- Read state on older items should not be lost when a new item arrives.

If the websocket reconnects, the hook can continue to reconcile with the REST payload as it does now.

## Components

### `web/components/notification/useNotifications.ts`

This remains the orchestration layer.

Responsibilities:

- Fetch initial notifications.
- Open and manage websocket updates.
- Handle optimistic `markAsRead` and `markAllAsRead` mutations.
- Roll back on failure.
- Expose state and handlers to the UI.

### `web/components/notification/NotificationDropdown.tsx`

This stays presentational.

Responsibilities:

- Render the notification icon.
- Toggle the panel open and closed.
- Pass the item-click and mark-all handlers into child components.

### `web/components/notification/NotificationList.tsx`

This stays a simple list renderer.

Responsibilities:

- Render each notification item.
- Forward click events to the hook-provided handler.

### `web/components/notification/NotificationItem.tsx`

This remains the clickable unit of interaction.

Responsibilities:

- Display title, message, time, and unread indicator.
- Trigger the single-item read handler on click.

## Data Flow

```text
Initial load
  -> useNotifications fetches /notifications
  -> list is normalized and sorted
  -> unread count is derived from state

Single notification click
  -> NotificationItem onClick(id)
  -> useNotifications marks one item read locally
  -> PUT /notifications/{id}/read
  -> rollback if the request fails

Read all button
  -> NotificationPanel onMarkAll()
  -> useNotifications marks all items read locally
  -> PUT /notifications/read-all
  -> rollback if the request fails

Websocket update
  -> NotificationWsClient receives live payload
  -> payload is normalized
  -> item is merged into state if it is new
```

## Error Handling

- Failed single-item read requests should restore the prior list and set an error message.
- Failed read-all requests should restore the prior list and set an error message.
- A failed initial notification fetch should keep the existing empty state and surface an error.
- API 404, 405, or 501 responses on fetch should still resolve to an empty list, matching the current fallback behavior.
- Websocket reconnect and receive failures should not block the UI.

## Edge Cases

- Clicking an already-read notification should not break the state model; it can safely be treated as a no-op from the user’s perspective.
- A notification arriving live while the user is marking others as read should still be merged without duplication.
- If the notification list is empty, the panel should still render a stable empty state.
- If the user closes the panel mid-mutation, the optimistic update should still complete or roll back correctly.

## Testing Strategy

### Unit tests

- `useNotifications` should mark a single item as read optimistically and call the single-read API.
- `useNotifications` should mark all items as read optimistically and call the bulk API.
- Failed mutations should restore the previous list.
- Websocket merges should not duplicate items.
- Reconciliation should preserve the latest state ordering.

### Component tests

- Clicking a notification item should invoke the single-read handler.
- Clicking the panel button should invoke the read-all handler.
- The unread badge should disappear when all notifications are read.

### Manual verification

- Open the panel with unread notifications.
- Click one item and confirm only that item becomes read.
- Click the read-all button and confirm every item becomes read.
- Confirm live websocket updates still appear while the panel is open.

## Implementation Notes

- Keep the optimistic update logic in the hook rather than pushing API calls into presentational components.
- Preserve the existing fetch/reconnect behavior instead of adding a refetch after every mutation.
- Avoid adding a second read-all entry point unless a future UX requirement needs it.

## Success Criteria

- Clicking a notification item marks only that notification as read.
- The existing panel button marks all notifications as read.
- The unread count updates immediately in the UI.
- Failed mutations roll back cleanly.
- Live notifications still arrive and merge without duplication.
