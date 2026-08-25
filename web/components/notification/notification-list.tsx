'use client';

import React from 'react';
import type { Notification } from '@/lib/types/notification.types';
import NotificationItem from '@/components/notification/notification-item';

export default function NotificationList({
  items,
  onItemClick,
}: {
  items: Notification[];
  onItemClick: (id: string) => void;
}) {
  if (items.length === 0) {
    return <p className="meta px-4 py-8 text-center">Nothing new</p>;
  }

  return (
    <div className="max-h-80 overflow-y-auto">
      {items.map((it) => (
        <NotificationItem key={it.id} item={it} onClick={onItemClick} />
      ))}
    </div>
  );
}
