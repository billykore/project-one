'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { ApiError } from '@/lib/api';
import {
  fetchNotifications,
  markAllNotificationsAsRead,
  markNotificationAsRead,
} from '@/lib/notifications-api';
import { NotificationWsClient } from '@/lib/notifications-ws';
import type { WsConnectionState } from '@/lib/notifications-ws';
import type { Notification } from '../../lib/types/notification.types';

/**
 * Merge an incoming notification into the existing list:
 * - Deduplicate by id (keep existing entry if id already present)
 * - Prepend the new item
 * - Keep the list sorted descending by createdAt
 */
function mergeNotification(prev: Notification[], incoming: Notification): Notification[] {
  if (prev.some((n) => n.id === incoming.id)) return prev;
  return [incoming, ...prev].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
  );
}

/**
 * Merge a REST refresh result into the current live state:
 * - Keep unread status from live state for ids that already exist
 * - Add any new items from the REST response
 * - Re-sort by createdAt descending
 */
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
  return Array.from(liveMap.values()).sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
  );
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

  // Stable ref so the WS client callback can access latest state
  // without being included in its own dependency array.
  const setNotificationsRef = useRef(setNotifications);

  useEffect(() => {
    setNotificationsRef.current = setNotifications;
    let isActive = true;

    const initialLoad = async () => {
      await loadNotifications();
      if (isActive) {
        setIsLoading(false);
      }
    };

    void initialLoad();

    // Attach WebSocket stream after initial REST load.
    const client = new NotificationWsClient();

    // On reconnect after a disconnect, fetch REST history to recover any
    // notifications that arrived while the socket was down, then reconcile
    // with the current in-memory live state to avoid duplicates.
    client.onReconnect = () => {
      fetchNotifications()
        .then((fresh) => {
          if (isActive) {
            setNotificationsRef.current((live) => reconcileNotifications(live, fresh));
          }
        })
        .catch(() => {
          // Silently ignore reconciliation errors — live data is still intact.
        });
    };

    client.onStateChange(setConnectionState);

    const unsub = client.onMessage((notification) => {
      setNotificationsRef.current((prev) => mergeNotification(prev, notification));
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
    connectionState,
    refresh,
    toggleDropdown,
    markAllAsRead,
    markAsRead,
    setNotifications,
  } as const;
}
