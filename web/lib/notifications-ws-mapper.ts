import type { Notification, NotificationType } from './types/notification.types';
import type { WsNotificationPayload } from './notifications-ws';

const ALLOWED_TYPES = ['success', 'info', 'warning', 'error'] as const;

export function mapWsPayload(raw: unknown): Notification | null {
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) return null;

  const p = raw as Partial<WsNotificationPayload>;
  if (p.id === undefined || p.id === null) return null;

  const actor = (p.actor_username ?? '').trim() || 'Someone';
  const fallback: Pick<Notification, 'title' | 'message'> =
    p.type === 'follow'
      ? { title: 'New follower', message: `${actor} started following you.` }
      : p.type === 'like'
        ? { title: 'Post liked', message: `${actor} liked your post.` }
        : p.type === 'comment'
          ? { title: 'New comment', message: `${actor} commented on your post.` }
          : { title: 'Notification', message: `${actor} sent you an update.` };

  const createdAt = p.created_at ?? new Date().toISOString();
  const type: NotificationType =
    p.type && ALLOWED_TYPES.includes(p.type as (typeof ALLOWED_TYPES)[number])
      ? (p.type as NotificationType)
      : 'info';

  return {
    id: String(p.id),
    title: (p.title ?? '').trim() || fallback.title,
    message: (p.body ?? '').trim() || fallback.message,
    type,
    isRead: p.is_read ?? false,
    createdAt,
  };
}
