/**
 * Mapper from WebSocket notification stream payload to the frontend Notification model.
 *
 * The backend streams dto.NotificationResponse as JSON with these fields:
 *   id, user_id, actor_id, actor_username, type, post_id?, comment_id?,
 *   is_read, created_at, title?, body?
 *
 * We reuse the same normalisation semantics as notifications-api.ts so that
 * live-streamed and REST-fetched notifications look identical in the UI.
 */

import type { Notification, NotificationType } from './types/notification.types';
import type { WsNotificationPayload } from './notifications-ws';

const ALLOWED_TYPES: NotificationType[] = ['success', 'info', 'warning', 'error'];

function normalizeType(type: string | undefined): NotificationType {
  if (!type) return 'info';
  if (ALLOWED_TYPES.includes(type as NotificationType)) {
    return type as NotificationType;
  }
  return 'info';
}

function buildFallbackContent(
  type: string | undefined,
  actor: string,
): Pick<Notification, 'title' | 'message'> {
  switch (type) {
    case 'follow':
      return { title: 'New follower', message: `${actor} started following you.` };
    case 'like':
      return { title: 'Post liked', message: `${actor} liked your post.` };
    case 'comment':
      return { title: 'New comment', message: `${actor} commented on your post.` };
    default:
      return { title: 'Notification', message: `${actor} sent you an update.` };
  }
}

/**
 * Map a raw WS message payload to a Notification.
 * Returns null when the payload is not a valid notification object.
 */
export function mapWsPayload(raw: unknown): Notification | null {
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) return null;

  const p = raw as Partial<WsNotificationPayload>;

  if (p.id === undefined || p.id === null) return null;

  const actor = (p.actor_username ?? '').trim() || 'Someone';
  const fallback = buildFallbackContent(p.type, actor);
  const createdAt = p.created_at ?? new Date().toISOString();

  return {
    id: String(p.id),
    title: (p.title ?? '').trim() || fallback.title,
    message: (p.body ?? '').trim() || fallback.message,
    type: normalizeType(p.type),
    isRead: p.is_read ?? false,
    createdAt,
  };
}
