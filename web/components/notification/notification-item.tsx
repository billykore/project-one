'use client';

import React from 'react';
import type { Notification } from '@/lib/types/notification.types';
import { formatTimestamp } from '@/lib/utils/date.utils';

export default function NotificationItem({
  item,
  onClick,
}: {
  item: Notification;
  onClick: (id: string) => void;
}) {
  return (
    <button
      type="button"
      aria-label={`Notification ${item.title}`}
      onClick={() => onClick(item.id)}
      className="relative block w-full border-b border-rule px-4 py-3 text-left transition-colors last:border-b-0 hover:bg-sunken"
    >
      {/* Unread entries keep their mark in the margin until you read them. */}
      <span
        aria-hidden="true"
        className={`absolute inset-y-0 left-0 w-[3px] ${item.isRead ? 'bg-transparent' : 'bg-ember'}`}
      />
      <div className="flex items-baseline justify-between gap-3">
        <span className={`truncate text-sm ${item.isRead ? 'font-medium text-muted' : 'font-semibold text-ink'}`}>
          {item.title}
        </span>
        <time className="meta shrink-0">{formatTimestamp(item.createdAt)}</time>
      </div>
      <p className="prose-lead mt-1 line-clamp-2 text-sm">{item.message}</p>
    </button>
  );
}
