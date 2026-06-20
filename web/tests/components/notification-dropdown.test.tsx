import React from 'react';
import { act } from 'react';
import { createRoot } from 'react-dom/client';
import { describe, expect, it, vi } from 'vitest';
import NotificationDropdown from '@/components/notification/notification-dropdown';

const markAsRead = vi.fn();
const markAllAsRead = vi.fn();

vi.mock('@/hooks/use-notifications', () => ({
  useNotifications: () => ({
    notifications: [
      {
        id: '1',
        title: 'New comment',
        message: 'Someone replied',
        type: 'info',
        isRead: false,
        createdAt: '2026-06-19T12:00:00.000Z',
      },
    ],
    isOpen: true,
    unreadCount: 1,
    connectionState: 'connected',
    toggleDropdown: vi.fn(),
    markAllAsRead,
    markAsRead,
  }),
}));

describe('NotificationDropdown', () => {
  it('forwards item clicks and read-all clicks to the hook handlers', async () => {
    const container = document.createElement('div');
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(<NotificationDropdown />);
    });

    const item = container.querySelector('button[aria-label="Notification New comment"]');
    const readAll = Array.from(container.querySelectorAll('button')).find(
      (button) => button.textContent?.includes('Mark all as read'),
    );

    item?.dispatchEvent(new MouseEvent('click', { bubbles: true }));
    readAll?.dispatchEvent(new MouseEvent('click', { bubbles: true }));

    expect(markAsRead).toHaveBeenCalledWith('1');
    expect(markAllAsRead).toHaveBeenCalledTimes(1);
  });
});