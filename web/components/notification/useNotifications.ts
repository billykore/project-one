'use client';

import { useMemo, useState } from 'react';
import type { Notification } from '../../lib/types/notification.types';

const mockNotifications: Notification[] = [
  {
    id: '1',
    title: 'Welcome!',
    message: 'Thanks for joining Project One.',
    type: 'info',
    isRead: false,
    createdAt: new Date(Date.now() - 1000 * 60 * 5).toISOString(),
  },
  {
    id: '2',
    title: 'New follower',
    message: 'alex started following you.',
    type: 'success',
    isRead: false,
    createdAt: new Date(Date.now() - 1000 * 60 * 60).toISOString(),
  },
  {
    id: '3',
    title: 'System',
    message: 'Planned maintenance tonight.',
    type: 'warning',
    isRead: true,
    createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(),
  },
];

export function useNotifications() {
  const [notifications, setNotifications] = useState<Notification[]>(mockNotifications);
  const [isOpen, setIsOpen] = useState(false);

  const unreadCount = useMemo(() => notifications.filter((n) => !n.isRead).length, [notifications]);

  function toggleDropdown() {
    setIsOpen((s) => !s);
  }

  function markAllAsRead() {
    setNotifications((prev) => prev.map((n) => ({ ...n, isRead: true })));
  }

  function markAsRead(id: string) {
    setNotifications((prev) => prev.map((n) => (n.id === id ? { ...n, isRead: true } : n)));
  }

  return {
    notifications,
    isOpen,
    unreadCount,
    toggleDropdown,
    markAllAsRead,
    markAsRead,
    setNotifications,
  } as const;
}
