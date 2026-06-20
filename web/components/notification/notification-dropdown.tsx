'use client';

import React, { useEffect, useRef } from 'react';
import { useNotifications } from '@/hooks/use-notifications';
import NotificationIcon from '@/components/notification/notification-icon';
import NotificationPanel from '@/components/notification/notification-panel';
import NotificationList from '@/components/notification/notification-list';

export default function NotificationDropdown() {
  const { notifications, isOpen, unreadCount, connectionState, toggleDropdown, markAllAsRead, markAsRead } = useNotifications();
  const containerRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (!containerRef.current) return;
      if (e.target instanceof Node && !containerRef.current.contains(e.target)) {
        if (isOpen) toggleDropdown();
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen, toggleDropdown]);

  return (
    <div className="relative" ref={containerRef}>
      <NotificationIcon unreadCount={unreadCount} isOpen={isOpen} onClick={toggleDropdown} />
      {isOpen && (
        <NotificationPanel onMarkAll={markAllAsRead} unreadCount={unreadCount} connectionState={connectionState}>
          <NotificationList items={notifications} onItemClick={markAsRead} />
        </NotificationPanel>
      )}
    </div>
  );
}
