'use client';

import React from 'react';
import type { Notification } from '../../lib/types/notification.types';
import { formatTimestamp } from '../../lib/utils/date.utils';

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
      className={`flex w-full gap-3 px-4 py-3 items-start text-left hover:bg-gray-50 dark:hover:bg-zinc-800 ${item.isRead ? 'opacity-80' : 'bg-white'}`}
    >
      <div className="flex-shrink-0">
        <div className="h-8 w-8 rounded-full bg-gray-200 dark:bg-zinc-700 flex items-center justify-center text-xs font-semibold text-gray-700 dark:text-zinc-200">
          {item.title.slice(0, 1).toUpperCase()}
        </div>
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between gap-2">
          <div className="text-sm font-semibold truncate text-gray-900 dark:text-zinc-50">{item.title}</div>
          <div className="text-xs text-gray-500 dark:text-zinc-400">{formatTimestamp(item.createdAt)}</div>
        </div>
        <div className="mt-1 text-sm text-gray-600 dark:text-zinc-300 line-clamp-2">{item.message}</div>
      </div>
      {!item.isRead && <div className="h-2 w-2 rounded-full bg-indigo-500 mt-3" />}
    </button>
  );
}
