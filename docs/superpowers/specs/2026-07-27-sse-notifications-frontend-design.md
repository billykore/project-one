# SSE Notifications Frontend Design

## Goal

Replace the frontend WebSocket notification stream with Server-Sent Events while keeping the existing notification UI, REST history loading, deduplication, and optimistic read behavior.

## Decisions

- Use a same-origin Next.js proxy route: `GET /api/notifications/stream`.
- Fully replace the WebSocket client; do not keep a WebSocket fallback.
- Keep notification state ownership in `useNotifications`.
- Reuse `normalizeNotification` so REST and live payloads keep one shape.

## Architecture

Add a Next.js route handler at `web/app/api/notifications/stream/route.ts`. It forwards the incoming request cookies to the Go backend endpoint `GET /notifications/stream` and returns the backend response as raw `text/event-stream`. This route must not use the existing JSON proxy helper because SSE responses are long-lived streams, not JSON payloads.

Add `web/lib/notifications-sse.ts`, update the consuming hook and tests to import it, then delete `web/lib/notifications-ws.ts` once no callers remain. The consuming hook should import a `NotificationSseClient` and a generic connection state type:

```ts
type ConnectionState = "connected" | "reconnecting" | "offline";
```

The SSE client is responsible only for transport:

- Open `new EventSource("/api/notifications/stream")`.
- Listen for `notification` events.
- Parse `event.data` as JSON.
- Normalize the payload with `normalizeNotification`.
- Notify the hook through `onMessage`.
- Emit connection state changes.
- Close the `EventSource` on disconnect.

`useNotifications` remains responsible for:

- Initial REST load from `/api/notifications`.
- Merging live notifications by `id`.
- Sorting newest first.
- Re-fetching REST notifications after reconnect to fill missed events.
- Optimistic mark-one-read and mark-all-read rollback behavior.

## Data Flow

1. A signed-in user loads the app.
2. `useNotifications()` fetches historical notifications through `GET /api/notifications`.
3. `NotificationSseClient` opens `GET /api/notifications/stream` with `EventSource`.
4. The Next.js route forwards cookies to the Go backend `GET /notifications/stream`.
5. The backend keeps the stream open and emits `event: notification` with JSON data.
6. The SSE client parses and normalizes each event.
7. `useNotifications` deduplicates by notification `id`, sorts newest first, and updates the dropdown/panel.
8. If the stream reconnects after an interruption, `useNotifications` fetches REST notifications and reconciles any missed items.

## Error Handling

Browser-native `EventSource` retry handles reconnect timing. The client should avoid custom backoff unless the native behavior proves insufficient.

- `onopen`: set state to `connected`. If the previous state was `reconnecting`, call the existing reconnect refresh callback.
- `onerror`: set state to `reconnecting` while the client is still active.
- invalid JSON event data: ignore the event.
- unrecognized notification payload: ignore the event.
- `disconnect()`: close the stream, clear handlers, and set state to `offline`.

Authentication failures surface as stream errors. The UI should continue showing REST-loaded notifications and an offline/reconnecting indicator rather than clearing notification state.

## Testing

Update existing hook tests to mock the SSE client instead of the WebSocket client.

Add the smallest useful transport coverage:

- parses a valid `notification` event and calls the message handler;
- ignores invalid JSON;
- emits `connected`, `reconnecting`, and `offline` states.

No new npm dependency is required.

## Out of Scope

- Notification UI redesign.
- Browser push notifications.
- Cross-tab shared SSE connection.
- Custom retry/backoff tuning.
- WebSocket fallback.
