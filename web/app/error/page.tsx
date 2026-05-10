'use client';

import { useSearchParams } from 'next/navigation';
import { Suspense } from 'react';
import ErrorDisplay from '@/components/layout/error';

function ErrorContent() {
  const searchParams = useSearchParams();
  const message = searchParams.get('message') || undefined;

  return (
    <ErrorDisplay 
      message={message} 
      onRetry={() => window.location.reload()} 
    />
  );
}

export default function ErrorPage() {
  return (
    <main className="flex flex-1 items-center justify-center bg-zinc-50 dark:bg-black p-4">
      <Suspense fallback={
        <div className="flex flex-col items-center justify-center min-h-[60vh]">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-zinc-900 dark:border-zinc-50"></div>
        </div>
      }>
        <ErrorContent />
      </Suspense>
    </main>
  );
}
