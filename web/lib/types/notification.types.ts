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
