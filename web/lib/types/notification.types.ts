export type NotificationType = 'success' | 'info' | 'warning' | 'error';

export interface Notification {
  id: string;
  title: string;
  message: string;
  type: NotificationType;
  isRead: boolean;
  createdAt: string; // ISO date string for simplicity
  avatar?: string;
  link?: string;
}

export interface BackendNotification {
  id: string | number;
  user_id?: number;
  actor_id?: number;
  actor_name?: string;
  actor_username?: string;
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
}
