import { api } from "./api";
import type { Notification, NotificationType } from "./types/notification.types";

type BackendNotification = {
  id: string | number;
  actor_name?: string;
  post_id?: number;
  comment_id?: number;
  title?: string;
  message?: string;
  body?: string;
  type?: string;
  is_read?: boolean;
  isRead?: boolean;
  created_at?: string;
  createdAt?: string;
  avatar?: string;
  link?: string;
  url?: string;
};

type NotificationsListResponse = {
  notifications?: BackendNotification[];
  data?: BackendNotification[];
  items?: BackendNotification[];
};

const NOTIFICATIONS_ENDPOINT = "/api/v1/notifications";

const ALLOWED_TYPES = ["success", "info", "warning", "error"] as const;

function normalizeNotification(item: BackendNotification): Notification {
  const createdAt = item.createdAt ?? item.created_at ?? new Date().toISOString();
  const actor = item.actor_name?.trim() || "Someone";
  const fallback: Pick<Notification, "title" | "message"> =
    item.type === "follow"
      ? { title: "New follower", message: `${actor} started following you.` }
      : item.type === "like"
        ? { title: "Post liked", message: `${actor} liked your post.` }
        : item.type === "comment"
          ? { title: "New comment", message: `${actor} commented on your post.` }
          : { title: "Notification", message: `${actor} sent you an update.` };
  const type: NotificationType =
    item.type && ALLOWED_TYPES.includes(item.type as (typeof ALLOWED_TYPES)[number])
      ? (item.type as NotificationType)
      : "info";

  return {
    id: String(item.id),
    title: item.title?.trim() || fallback.title,
    message: item.message?.trim() || item.body?.trim() || fallback.message,
    type,
    isRead: item.isRead ?? item.is_read ?? false,
    createdAt,
    avatar: item.avatar,
    link: item.link ?? item.url,
  };
}

export async function fetchNotifications(): Promise<Notification[]> {
  const payload = await api.get<NotificationsListResponse | BackendNotification[]>(NOTIFICATIONS_ENDPOINT);
  const notifications = (Array.isArray(payload)
    ? payload
    : payload.notifications ?? payload.data ?? payload.items ?? []
  ).map(normalizeNotification);

  return notifications.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
}

export async function markNotificationAsRead(id: string): Promise<void> {
  await api.put(`${NOTIFICATIONS_ENDPOINT}/${id}/read`, {});
}

export async function markAllNotificationsAsRead(): Promise<void> {
  await api.put(`${NOTIFICATIONS_ENDPOINT}/read-all`, {});
}