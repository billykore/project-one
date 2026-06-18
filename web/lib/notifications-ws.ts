/**
 * WebSocket notification client for real-time notification streaming.
 *
 * Transport contract:
 *   Endpoint : ws(s)://<host>/websocket  (configured via NEXT_PUBLIC_WS_URL)
 *   Protocol : native WebSocket (no sub-protocol negotiation)
 *   Auth     : The backend handler at /websocket currently validates the JWT from the
 *              "Authorization: ******" HTTP header.  Native browser WebSockets
 *              do NOT allow setting arbitrary headers during the upgrade handshake.
 *              The browser DOES automatically forward cookies that belong to the WS
 *              host on the upgrade request.  For this client to authenticate, the
 *              backend /websocket handler must be updated to also accept the
 *              "access_token" cookie (matching the behaviour of the REST auth
 *              middleware in internal/api/middleware/authorization.go).
 *              Until that backend change lands, connections will be rejected with 401.
 *   Messages : JSON-encoded NotificationResponse (see WsNotificationPayload below).
 */

import type { Notification } from './types/notification.types';
import { mapWsPayload } from './notifications-ws-mapper';

export type WsConnectionState = 'connected' | 'reconnecting' | 'offline';

export type WsNotificationPayload = {
  id: number;
  user_id: number;
  actor_id?: number;
  actor_username?: string;
  type?: string;
  post_id?: number;
  comment_id?: number;
  is_read?: boolean;
  created_at?: string;
  title?: string;
  body?: string;
};

type MessageHandler = (notification: Notification) => void;
type StateHandler = (state: WsConnectionState) => void;

const WS_URL =
  process.env.NEXT_PUBLIC_WS_URL ?? 'ws://localhost:8080/websocket';

/** Minimum reconnect delay in ms. */
const BACKOFF_BASE_MS = 1_000;
/** Maximum reconnect delay in ms (30 s). */
const BACKOFF_CAP_MS = 30_000;
/** Maximum reconnect attempts before giving up entirely. */
const MAX_ATTEMPTS = 10;

const DEBUG = process.env.NEXT_PUBLIC_WS_DEBUG === 'true';

function dbg(...args: unknown[]) {
  if (DEBUG) {
    // eslint-disable-next-line no-console
    console.debug('[notifications-ws]', ...args);
  }
}

/**
 * NotificationWsClient manages the WebSocket lifecycle for notification streaming.
 *
 * Usage:
 *   const client = new NotificationWsClient();
 *   client.onMessage(n => setState(prev => [n, ...prev]));
 *   client.onStateChange(s => setConnectionState(s));
 *   client.connect();
 *   // On cleanup:
 *   client.disconnect();
 */
export class NotificationWsClient {
  private ws: WebSocket | null = null;
  private attempts = 0;
  private retryTimer: ReturnType<typeof setTimeout> | null = null;
  private explicitClose = false;
  private messageHandlers: Set<MessageHandler> = new Set();
  private stateHandlers: Set<StateHandler> = new Set();
  private _state: WsConnectionState = 'offline';
  /** Called when a clean reconnect succeeds after previous failure. */
  onReconnect: (() => void) | null = null;

  get state(): WsConnectionState {
    return this._state;
  }

  private setState(next: WsConnectionState) {
    if (this._state === next) return;
    this._state = next;
    dbg('state →', next);
    this.stateHandlers.forEach((h) => h(next));
  }

  onMessage(handler: MessageHandler): () => void {
    this.messageHandlers.add(handler);
    return () => this.messageHandlers.delete(handler);
  }

  onStateChange(handler: StateHandler): () => void {
    this.stateHandlers.add(handler);
    return () => this.stateHandlers.delete(handler);
  }

  connect() {
    if (this.explicitClose) return;
    if (this.ws && this.ws.readyState <= WebSocket.OPEN) return;

    dbg('connecting to', WS_URL, 'attempt', this.attempts + 1);

    try {
      this.ws = new WebSocket(WS_URL);
    } catch (err) {
      dbg('WebSocket constructor error', err);
      this.scheduleReconnect();
      return;
    }

    this.ws.onopen = () => {
      dbg('connected');
      const wasReconnect = this.attempts > 0;
      this.attempts = 0;
      this.setState('connected');
      if (wasReconnect && this.onReconnect) {
        this.onReconnect();
      }
    };

    this.ws.onmessage = (event: MessageEvent) => {
      dbg('message', event.data);
      let payload: unknown;
      try {
        payload = JSON.parse(event.data as string) as unknown;
      } catch {
        dbg('failed to parse message JSON, skipping');
        return;
      }
      const notification = mapWsPayload(payload);
      if (!notification) {
        dbg('payload did not map to a notification, skipping');
        return;
      }
      this.messageHandlers.forEach((h) => h(notification));
    };

    this.ws.onerror = (event) => {
      dbg('error', event);
    };

    this.ws.onclose = (event) => {
      dbg('closed code=%d reason=%s', event.code, event.reason);
      this.ws = null;
      if (this.explicitClose) {
        this.setState('offline');
        return;
      }
      this.scheduleReconnect();
    };
  }

  private scheduleReconnect() {
    if (this.explicitClose) return;
    if (this.attempts >= MAX_ATTEMPTS) {
      dbg('max reconnect attempts reached, giving up');
      this.setState('offline');
      return;
    }

    this.setState('reconnecting');
    this.attempts += 1;
    const jitter = Math.random() * 500;
    const delay = Math.min(BACKOFF_BASE_MS * 2 ** (this.attempts - 1), BACKOFF_CAP_MS) + jitter;
    dbg('reconnecting in', delay.toFixed(0), 'ms (attempt', this.attempts, ')');

    this.retryTimer = setTimeout(() => {
      this.retryTimer = null;
      this.connect();
    }, delay);
  }

  disconnect() {
    dbg('disconnect called');
    this.explicitClose = true;
    if (this.retryTimer !== null) {
      clearTimeout(this.retryTimer);
      this.retryTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.setState('offline');
  }

  /** Reset the client so it can be reconnected after an explicit disconnect. */
  reset() {
    this.explicitClose = false;
    this.attempts = 0;
  }
}
