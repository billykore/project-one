'use client';

import React from 'react';

export default function NotificationIcon({
  unreadCount,
  isOpen,
  onClick,
}: {
  unreadCount: number;
  isOpen: boolean;
  onClick: () => void;
}) {
  return (
    <button
      aria-label="Notifications"
      aria-haspopup="true"
      aria-expanded={isOpen}
      onClick={onClick}
      className="relative h-9 w-9 flex items-center justify-center rounded-full hover:bg-gray-100 dark:hover:bg-zinc-800 focus:outline-none"
    >
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" className="h-5 w-5 text-gray-700 dark:text-zinc-200">
        <path d="M12 2a6 6 0 00-6 6v3.586L4.293 14.293A1 1 0 005 16h14a1 1 0 00.707-1.707L18 11.586V8a6 6 0 00-6-6z" fill="currentColor" />
        <path d="M9.293 18.707A2 2 0 0011 20h2a2 2 0 001.707-1.293" stroke="currentColor" strokeWidth="0" />
      </svg>

      {unreadCount > 0 && (
        <span className="absolute -top-0.5 -right-0.5 inline-flex items-center justify-center rounded-full bg-red-500 text-white text-[10px] h-5 w-5">
          {unreadCount}
        </span>
      )}
    </button>
  );
}
