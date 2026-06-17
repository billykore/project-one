'use client';

import React from 'react';
import type { Notification } from '../../lib/types/notification.types';
import NotificationItem from './NotificationItem';

export default function NotificationList({
  items,
  onItemClick,
}: {
  items: Notification[];
  onItemClick: (id: string) => void;
}) {
  if (items.length === 0) {
    return <div className="p-4 text-center text-sm text-gray-500 dark:text-zinc-400">No notifications</div>;
  }

  return (
    <div className="max-h-80 overflow-y-auto">
      {items.map((it) => (
        <NotificationItem key={it.id} item={it} onClick={onItemClick} />
      ))}
    </div>
  );
}
