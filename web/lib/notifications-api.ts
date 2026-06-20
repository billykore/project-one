import { api } from "./api";
import type { Notification, NotificationType, BackendNotification } from "./types/notification.types";

type NotificationsListResponse = {
  notifications?: BackendNotification[];
  data?: BackendNotification[];
  items?: BackendNotification[];
};

const NOTIFICATIONS_ENDPOINT = "/api/v1/notifications";

const ALLOWED_TYPES = ["success", "info", "warning", "error"] as const;

export function normalizeNotification(item: unknown): Notification | null {
  if (!item || typeof item !== "object" || Array.isArray(item)) return null;
  const p = item as BackendNotification;
  if (p.id === undefined || p.id === null) return null;

  const createdAt = p.createdAt ?? p.created_at ?? new Date().toISOString();
  // ponytail: handle both HTTP (actor_name) and WebSocket (actor_username) actor fields
  const actor = p.actor_name?.trim() || p.actor_username?.trim() || "Someone";
  const fallback: Pick<Notification, "title" | "message"> =
    p.type === "follow"
      ? { title: "New follower", message: `${actor} started following you.` }
      : p.type === "like"
        ? { title: "Post liked", message: `${actor} liked your post.` }
        : p.type === "comment"
          ? { title: "New comment", message: `${actor} commented on your post.` }
          : { title: "Notification", message: `${actor} sent you an update.` };

  const type: NotificationType =
    p.type && ALLOWED_TYPES.includes(p.type as (typeof ALLOWED_TYPES)[number])
      ? (p.type as NotificationType)
      : "info";

  return {
    id: String(p.id),
    title: p.title?.trim() || fallback.title,
    message: p.message?.trim() || p.body?.trim() || fallback.message,
    type,
    isRead: p.isRead ?? p.is_read ?? false,
    createdAt,
    avatar: p.avatar,
    link: p.link ?? p.url,
  };
}

export async function fetchNotifications(): Promise<Notification[]> {
  const payload = await api.get<NotificationsListResponse | BackendNotification[]>(NOTIFICATIONS_ENDPOINT);
  const notifications = (Array.isArray(payload)
    ? payload
    : payload.notifications ?? payload.data ?? payload.items ?? []
  )
    .map(normalizeNotification)
    .filter((n): n is Notification => n !== null);

  return notifications.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
}

export async function markNotificationAsRead(id: string): Promise<void> {
  await api.put(`${NOTIFICATIONS_ENDPOINT}/${id}/read`, {});
}

export async function markAllNotificationsAsRead(): Promise<void> {
  await api.put(`${NOTIFICATIONS_ENDPOINT}/read-all`, {});
}