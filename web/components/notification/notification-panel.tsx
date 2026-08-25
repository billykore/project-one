'use client';

import React from 'react';
import type { ConnectionState } from '@/lib/notifications-sse';

const CONNECTION: Record<ConnectionState | 'offline', { label: string; dot: string; pulse: boolean }> = {
  connected: { label: 'Live', dot: 'bg-accent', pulse: false },
  reconnecting: { label: 'Reconnecting', dot: 'bg-sand', pulse: true },
  offline: { label: 'Offline', dot: 'bg-faint', pulse: false },
};

function ConnectionIndicator({ state }: { state: ConnectionState }) {
  const { label, dot, pulse } = CONNECTION[state] ?? CONNECTION.offline;
  return (
    <span className="meta flex items-center gap-1.5">
      <span className={`h-1.5 w-1.5 rounded-full ${dot} ${pulse ? 'animate-pulse' : ''}`} aria-hidden="true" />
      {label}
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
  connectionState?: ConnectionState;
}) {
  return (
    <div className="absolute right-0 z-50 mt-2 w-80 sm:w-96">
      <div className="pop enter-pop overflow-hidden">
        <div className="flex items-center justify-between border-b border-rule px-4 py-3">
          <h2 className="text-sm font-semibold text-ink">Notifications</h2>
          <ConnectionIndicator state={connectionState} />
        </div>
        <div>{children}</div>
        {unreadCount > 0 && (
          <div className="border-t border-rule p-2">
            <button onClick={onMarkAll} className="btn btn-ghost btn-block">
              Mark all as read
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
