import React from 'react';
import { act } from 'react';
import { createRoot } from 'react-dom/client';
import { describe, expect, it } from 'vitest';
import {
  NavbarActionsSkeleton,
  PageSkeletonLayout,
  PostsGridSkeleton,
} from '@/components/ui/page-skeleton';

describe('PageSkeleton components', () => {
  it('renders the page layout title and child content', async () => {
    const container = document.createElement('div');
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <PageSkeletonLayout title="Posts" rightActions={<NavbarActionsSkeleton />}>
          <div>content-marker</div>
        </PageSkeletonLayout>,
      );
    });

    expect(container.textContent).toContain('Posts');
    expect(container.textContent).toContain('content-marker');
  });

  it('renders the requested number of post skeleton cards', async () => {
    const container = document.createElement('div');
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(<PostsGridSkeleton count={3} />);
    });

    const cards = Array.from(container.querySelectorAll('div')).filter((node) => {
      const className = node.getAttribute('class') ?? '';
      return className.includes('rounded-2xl bg-white p-6 shadow-sm');
    });
    expect(cards).toHaveLength(3);
  });
});
