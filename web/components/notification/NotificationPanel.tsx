'use client';

import React from 'react';

export default function NotificationPanel({
  children,
  onMarkAll,
  unreadCount,
}: {
  children: React.ReactNode;
  onMarkAll: () => void;
  unreadCount: number;
}) {
  return (
    <div className="absolute right-0 mt-2 w-80 sm:w-96 z-50">
      <div className="rounded-lg bg-white dark:bg-zinc-900 border border-gray-200 dark:border-zinc-700 shadow-lg overflow-hidden">
        <div className="flex items-center justify-between px-4 py-3 border-b border-gray-100 dark:border-zinc-800">
          <div className="font-semibold text-gray-900 dark:text-zinc-50">Notifications</div>
          {unreadCount > 0 && (
            <button onClick={onMarkAll} className="text-xs px-2 py-1 rounded bg-gray-100 dark:bg-zinc-800 text-gray-700 dark:text-zinc-200">
              Mark all as read
            </button>
          )}
        </div>
        <div>{children}</div>
        {unreadCount > 0 && (
          <div className="p-3 border-t border-gray-100 dark:border-zinc-800">
            <button onClick={onMarkAll} className="w-full rounded-md bg-indigo-600 text-white px-3 py-2 text-sm">
              Mark all as read
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
