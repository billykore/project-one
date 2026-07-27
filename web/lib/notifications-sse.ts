import { normalizeNotification } from './notifications-api';
import type { Notification } from './types/notification.types';

export type ConnectionState = 'connected' | 'reconnecting' | 'offline';

type MessageHandler = (notification: Notification) => void;
type StateHandler = (state: ConnectionState) => void;

const STREAM_URL = '/api/notifications/stream';

export class NotificationSseClient {
  private source: EventSource | null = null;
  private messageHandler: MessageHandler | null = null;
  private stateHandler: StateHandler | null = null;
  private _state: ConnectionState = 'offline';
  onReconnect: (() => void) | null = null;

  get state(): ConnectionState {
    return this._state;
  }

  private setState(next: ConnectionState) {
    if (this._state === next) return;
    this._state = next;
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
    if (this.source) return;

    try {
      this.source = new EventSource(STREAM_URL);
    } catch {
      this.setState('offline');
      return;
    }

    this.source.addEventListener('notification', this.handleMessage);
    this.source.onmessage = this.handleMessage;
    this.source.onopen = () => {
      const wasReconnecting = this._state === 'reconnecting';
      this.setState('connected');
      if (wasReconnecting) {
        this.onReconnect?.();
      }
    };
    this.source.onerror = () => {
      this.setState('reconnecting');
    };
  }

  private handleMessage = (event: MessageEvent<string>) => {
    let payload: unknown;
    try {
      payload = JSON.parse(event.data) as unknown;
    } catch {
      return;
    }

    const notification = normalizeNotification(payload);
    if (!notification) return;

    this.messageHandler?.(notification);
  };

  disconnect() {
    this.source?.close();
    this.source = null;
    this.messageHandler = null;
    this.stateHandler = null;
    this.onReconnect = null;
    this.setState('offline');
  }
}
