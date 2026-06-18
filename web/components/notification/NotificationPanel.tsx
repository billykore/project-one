'use client';

import React from 'react';
import type { WsConnectionState } from '../../lib/notifications-ws';

function ConnectionIndicator({ state }: { state: WsConnectionState }) {
  if (state === 'connected') {
    return (
      <span className="flex items-center gap-1 text-xs text-green-600 dark:text-green-400">
        <span className="h-1.5 w-1.5 rounded-full bg-green-500" />
        Live
      </span>
    );
  }
  if (state === 'reconnecting') {
    return (
      <span className="flex items-center gap-1 text-xs text-amber-600 dark:text-amber-400">
        <span className="h-1.5 w-1.5 rounded-full bg-amber-500 animate-pulse" />
        Reconnecting
      </span>
    );
  }
  return (
    <span className="flex items-center gap-1 text-xs text-gray-400 dark:text-zinc-500">
      <span className="h-1.5 w-1.5 rounded-full bg-gray-400 dark:bg-zinc-500" />
      Offline
    </span>
  );
}

export default function NotificationPanel({
  children,
  onMarkAll,
  unreadCount,
  connectionState = 'offline',
}: {
  children: React.ReactNode;
  onMarkAll: () => void;
  unreadCount: number;
  connectionState?: WsConnectionState;
}) {
  return (
    <div className="absolute right-0 mt-2 w-80 sm:w-96 z-50">
      <div className="rounded-lg bg-white dark:bg-zinc-900 border border-gray-200 dark:border-zinc-700 shadow-lg overflow-hidden">
        <div className="flex items-center justify-between px-4 py-3 border-b border-gray-100 dark:border-zinc-800">
          <div className="flex items-center gap-2">
            <div className="font-semibold text-gray-900 dark:text-zinc-50">Notifications</div>
            <ConnectionIndicator state={connectionState} />
          </div>
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
