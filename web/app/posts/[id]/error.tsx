'use client';

import { useEffect } from 'react';
import ErrorDisplay from '@/components/layout/error';

export default function PostDetailError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error('Post detail page error:', error);
  }, [error]);

  return (
    <div className="flex min-h-screen flex-col bg-gray-50 dark:bg-gray-950">
      <main className="flex flex-1 items-center justify-center p-4">
        <ErrorDisplay 
          message={error.message || "Failed to load post. It might have been deleted or you may not have permission to view it."} 
          onRetry={() => reset()} 
          homePath="/posts"
        />
      </main>
    </div>
  );
}
