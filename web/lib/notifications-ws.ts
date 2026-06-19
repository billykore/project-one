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

const WS_URL = process.env.NEXT_PUBLIC_WS_URL ?? 'ws://localhost:8080/websocket';

const BACKOFF_BASE_MS = 1_000;
const BACKOFF_CAP_MS = 30_000;
const MAX_ATTEMPTS = 10;

const DEBUG = process.env.NEXT_PUBLIC_WS_DEBUG === 'true';

function dbg(...args: unknown[]) {
  if (DEBUG) {
    console.debug('[notifications-ws]', ...args);
  }
}

async function fetchWsToken(): Promise<string | null> {
  try {
    const res = await fetch('/api/ws-token', { credentials: 'include' });
    if (!res.ok) return null;
    const data = (await res.json()) as { token?: string | null };
    return data.token ?? null;
  } catch {
    dbg('failed to fetch WS token from /api/ws-token');
    return null;
  }
}

function buildWsUrl(token: string): string {
  const separator = WS_URL.includes('?') ? '&' : '?';
  return `${WS_URL}${separator}token=${encodeURIComponent(token)}`;
}

export class NotificationWsClient {
  private ws: WebSocket | null = null;
  private attempts = 0;
  private retryTimer: ReturnType<typeof setTimeout> | null = null;
  private explicitClose = false;
  private connecting = false;
  private messageHandler: MessageHandler | null = null;
  private stateHandler: StateHandler | null = null;
  private _state: WsConnectionState = 'offline';
  onReconnect: (() => void) | null = null;

  get state(): WsConnectionState {
    return this._state;
  }

  private setState(next: WsConnectionState) {
    if (this._state === next) return;
    this._state = next;
    dbg('state →', next);
    this.stateHandler?.(next);
  }

  onMessage(handler: MessageHandler): () => void {
    this.messageHandler = handler;
    return () => {
      if (this.messageHandler === handler) this.messageHandler = null;
    };
  }

  onStateChange(handler: StateHandler): () => void {
    this.stateHandler = handler;
    return () => {
      if (this.stateHandler === handler) this.stateHandler = null;
    };
  }

  connect() {
    if (this.explicitClose) return;
    if (this.ws && this.ws.readyState <= WebSocket.OPEN) return;
    if (this.connecting) return;

    this.connecting = true;

    void this.connectAsync().finally(() => {
      this.connecting = false;
    });
  }

  private async connectAsync() {
    if (this.explicitClose) return;

    const token = await fetchWsToken();
    if (this.explicitClose) return;

    if (!token) {
      dbg('no access_token available, skipping WebSocket connection');
      this.setState('offline');
      return;
    }

    const url = buildWsUrl(token);
    dbg('connecting to', WS_URL, 'attempt', this.attempts + 1);

    try {
      this.ws = new WebSocket(url);
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
      this.messageHandler?.(notification);
    };

    this.ws.onerror = (event) => {
      dbg('error', event);
    };

    this.ws.onclose = (event) => {
      dbg('closed code=%d reason=%s', event.code, event.reason);
      this.ws = null;
      if (this.explicitClose) return;
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
    this.messageHandler = null;
    this.stateHandler = null;
    this.setState('offline');
  }
}

