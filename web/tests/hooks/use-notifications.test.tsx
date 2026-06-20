import React from 'react';
import { act } from 'react';
import { createRoot } from 'react-dom/client';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useNotifications } from '@/hooks/use-notifications';

const mocks = vi.hoisted(() => ({
  fetchNotifications: vi.fn(),
  markNotificationAsRead: vi.fn(),
  markAllNotificationsAsRead: vi.fn(),
  // Vitest v4 propagates `new` to the implementation, so a regular function
  // (not an arrow function) is required for constructor mocks.
  NotificationWsClient: vi.fn().mockImplementation(function () {
    return {
      connect: vi.fn(),
      disconnect: vi.fn(),
      onMessage: vi.fn(() => vi.fn()),
      onReconnect: vi.fn(),
      onStateChange: vi.fn(),
    };
  }),
}));

vi.mock('@/lib/notifications-api', () => ({
  fetchNotifications: mocks.fetchNotifications,
  markNotificationAsRead: mocks.markNotificationAsRead,
  markAllNotificationsAsRead: mocks.markAllNotificationsAsRead,
}));

vi.mock('@/lib/notifications-ws', () => ({
  NotificationWsClient: mocks.NotificationWsClient,
}));

function Harness({ onReady }: { onReady: (api: ReturnType<typeof useNotifications>) => void }) {
  const api = useNotifications();

  React.useEffect(() => {
    onReady(api);
  }, [api, onReady]);

  return null;
}

describe('useNotifications', () => {
  afterEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });

  it('marks one notification read optimistically and keeps the update on success', async () => {
    mocks.fetchNotifications.mockResolvedValue([
      {
        id: '1',
        title: 'One',
        message: 'First',
        type: 'info',
        isRead: false,
        createdAt: '2026-06-19T10:00:00.000Z',
      },
    ]);
    mocks.markNotificationAsRead.mockResolvedValue(undefined);

    let api!: ReturnType<typeof useNotifications>;
    const root = createRoot(document.createElement('div'));

    await act(async () => {
      root.render(
        <Harness
          onReady={(next) => {
            api = next;
          }}
        />,
      );
    });

    await act(async () => {
      await api.markAsRead('1');
    });

    expect(mocks.markNotificationAsRead).toHaveBeenCalledWith('1');
    expect(api.notifications[0]?.isRead).toBe(true);
  });
});