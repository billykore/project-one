'use client';

import { useEffect } from 'react';
import ErrorDisplay from '@/components/layout/error';

export default function PostsError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error('Posts page error:', error);
  }, [error]);

  return (
    <div className="flex min-h-screen flex-col bg-gray-50 dark:bg-black">
      <main className="flex flex-1 items-center justify-center p-4">
        <ErrorDisplay 
          message={error.message || "Failed to load posts. Please make sure you are logged in."} 
          onRetry={() => reset()} 
          homePath="/home"
        />
      </main>
    </div>
  );
}
