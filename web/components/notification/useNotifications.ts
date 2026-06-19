'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { ApiError } from '@/lib/api';
import {
  fetchNotifications,
  markAllNotificationsAsRead,
  markNotificationAsRead,
} from '@/lib/notifications-api';
import { NotificationWsClient } from '@/lib/notifications-ws';
import type { WsConnectionState } from '@/lib/notifications-ws';
import type { Notification } from '../../lib/types/notification.types';

function sortByNewest(items: Notification[]): Notification[] {
  return [...items].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
}

function mergeNotification(prev: Notification[], incoming: Notification): Notification[] {
  if (prev.some((n) => n.id === incoming.id)) return prev;
  return sortByNewest([incoming, ...prev]);
}

function markOneRead(prev: Notification[], id: string): Notification[] {
  return prev.map((n) => (n.id === id ? { ...n, isRead: true } : n));
}

function markAllRead(prev: Notification[]): Notification[] {
  return prev.map((n) => ({ ...n, isRead: true }));
}

function reconcileNotifications(
  live: Notification[],
  fresh: Notification[],
): Notification[] {
  const liveMap = new Map(live.map((n) => [n.id, n]));
  for (const n of fresh) {
    if (!liveMap.has(n.id)) {
      liveMap.set(n.id, n);
    }
  }
  return sortByNewest(Array.from(liveMap.values()));
}

export function useNotifications() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [connectionState, setConnectionState] = useState<WsConnectionState>('offline');

  const unreadCount = useMemo(() => notifications.filter((n) => !n.isRead).length, [notifications]);

  const loadNotifications = useCallback(async () => {
    try {
      const nextNotifications = await fetchNotifications();
      setNotifications(nextNotifications);
      setError(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load notifications';
      setError(message);

      if (err instanceof ApiError && [404, 405, 501].includes(err.status)) {
        setNotifications([]);
      }
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

    const client = new NotificationWsClient();

    client.onReconnect = () => {
      fetchNotifications()
        .then((fresh) => {
          if (isActive) {
            setNotifications((live) => reconcileNotifications(live, fresh));
          }
        })
        .catch(() => {});
    };

    client.onStateChange(setConnectionState);

    const unsub = client.onMessage((notification) => {
      setNotifications((prev) => mergeNotification(prev, notification));
    });

    client.connect();

    return () => {
      isActive = false;
      unsub();
      client.disconnect();
    };
  }, [loadNotifications]);

  function toggleDropdown() {
    setIsOpen((s) => !s);
  }

  async function markAllAsRead() {
    let previous: Notification[] = [];

    setNotifications((current) => {
      previous = current;
      return markAllRead(current);
    });

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
    let previous: Notification[] = [];

    setNotifications((current) => {
      previous = current;
      return markOneRead(current, id);
    });

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
    connectionState,
    refresh,
    toggleDropdown,
    markAllAsRead,
    markAsRead,
    setNotifications,
  } as const;
}
