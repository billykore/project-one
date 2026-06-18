'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { ApiError } from '@/lib/api';
import {
  fetchNotifications,
  markAllNotificationsAsRead,
  markNotificationAsRead,
} from '@/lib/notifications-api';
import type { Notification } from '../../lib/types/notification.types';

export function useNotifications() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const unreadCount = useMemo(() => notifications.filter((n) => !n.isRead).length, [notifications]);

  const loadNotifications = useCallback(async () => {
    try {
      const nextNotifications = await fetchNotifications();
      setNotifications(nextNotifications);
      setError(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load notifications';
      setError(message);

      // If API is not available yet, keep UI functional with an empty list.
      if (err instanceof ApiError && [404, 405, 501].includes(err.status)) {
        setNotifications([]);
      }
    } finally {
      // no-op
    }
  }, []);

  const refresh = useCallback(async () => {
    try {
      setIsLoading(true);
      await loadNotifications();
    } finally {
      setIsLoading(false);
    }
  }, [loadNotifications]);

  useEffect(() => {
    let isActive = true;

    const initialLoad = async () => {
      await loadNotifications();
      if (isActive) {
        setIsLoading(false);
      }
    };

    void initialLoad();

    return () => {
      isActive = false;
    };
  }, [loadNotifications]);

  function toggleDropdown() {
    setIsOpen((s) => !s);
  }

  async function markAllAsRead() {
    const previous = notifications;
    setNotifications((prev) => prev.map((n) => ({ ...n, isRead: true })));

    try {
      await markAllNotificationsAsRead();
      setError(null);
    } catch (err) {
      setNotifications(previous);
      const message = err instanceof Error ? err.message : 'Failed to mark notifications as read';
      setError(message);
    }
  }

  async function markAsRead(id: string) {
    const previous = notifications;
    setNotifications((prev) => prev.map((n) => (n.id === id ? { ...n, isRead: true } : n)));

    try {
      await markNotificationAsRead(id);
      setError(null);
    } catch (err) {
      setNotifications(previous);
      const message = err instanceof Error ? err.message : 'Failed to mark notification as read';
      setError(message);
    }
  }

  return {
    notifications,
    isOpen,
    isLoading,
    error,
    unreadCount,
    refresh,
    toggleDropdown,
    markAllAsRead,
    markAsRead,
    setNotifications,
  } as const;
}
