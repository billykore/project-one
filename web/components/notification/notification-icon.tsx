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
      className={`relative flex h-8 w-8 items-center justify-center rounded-sm border transition-colors ${
        isOpen ? 'border-rule-strong bg-sunken' : 'border-transparent hover:bg-sunken'
      }`}
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="h-[18px] w-[18px] text-ink"
        aria-hidden="true"
      >
        <path d="M18 8A6 6 0 0 0 6 8c0 3.6-.9 5.3-1.9 6.4a.8.8 0 0 0 .6 1.3h14.6a.8.8 0 0 0 .6-1.3C18.9 13.3 18 11.6 18 8Z" />
        <path d="M10 19a2 2 0 0 0 4 0" />
      </svg>

      {unreadCount > 0 && (
        <span className="absolute -top-1 -right-1 inline-flex h-4 min-w-4 items-center justify-center rounded-xs bg-ember px-1 font-mono text-[0.6rem] leading-none font-semibold text-paper">
          {unreadCount > 99 ? '99+' : unreadCount}
        </span>
      )}
    </button>
  );
}
