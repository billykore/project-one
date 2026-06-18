import { api } from "./api";
import type { Notification, NotificationType } from "./types/notification.types";

type BackendNotification = {
  id: string | number;
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

const ALLOWED_TYPES: NotificationType[] = ["success", "info", "warning", "error"];

function normalizeType(type: string | undefined): NotificationType {
  if (!type) return "info";
  if (ALLOWED_TYPES.includes(type as NotificationType)) {
    return type as NotificationType;
  }
  return "info";
}

function normalizeNotification(item: BackendNotification): Notification {
  const createdAt = item.createdAt ?? item.created_at ?? new Date().toISOString();

  return {
    id: String(item.id),
    title: item.title?.trim() || "Notification",
    message: item.message?.trim() || item.body?.trim() || "",
    type: normalizeType(item.type),
    isRead: item.isRead ?? item.is_read ?? false,
    createdAt,
    avatar: item.avatar,
    link: item.link ?? item.url,
  };
}

function extractNotifications(payload: NotificationsListResponse | BackendNotification[]): BackendNotification[] {
  if (Array.isArray(payload)) {
    return payload;
  }

  return payload.notifications ?? payload.data ?? payload.items ?? [];
}

export async function fetchNotifications(): Promise<Notification[]> {
  const payload = await api.get<NotificationsListResponse | BackendNotification[]>(NOTIFICATIONS_ENDPOINT);
  const notifications = extractNotifications(payload).map(normalizeNotification);

  return notifications.sort((a, b) => {
    const aTime = new Date(a.createdAt).getTime();
    const bTime = new Date(b.createdAt).getTime();
    return bTime - aTime;
  });
}

export async function markNotificationAsRead(id: string): Promise<void> {
  await api.put(`${NOTIFICATIONS_ENDPOINT}/${id}/read`, {});
}

export async function markAllNotificationsAsRead(): Promise<void> {
  await api.put(`${NOTIFICATIONS_ENDPOINT}/read-all`, {});
}