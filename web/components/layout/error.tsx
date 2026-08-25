'use client';

import Link from 'next/link';

interface ErrorDisplayProps {
  message?: string;
  onRetry?: () => void;
  homePath?: string;
}

export default function ErrorDisplay({
  message = 'An unexpected error occurred.',
  onRetry,
  homePath = '/',
}: ErrorDisplayProps) {
  return (
    <div className="flex min-h-[60vh] items-center justify-center p-4">
      <div className="panel measure w-full border-t-2 border-t-ember p-8">
        <p className="meta">Request failed</p>
        <h1 className="mt-2 text-2xl font-semibold text-ink">That didn&apos;t go through</h1>
        <p className="prose-lead mt-3">{message}</p>
        <div className="mt-7 flex flex-wrap gap-2">
          {onRetry && (
            <button onClick={onRetry} className="btn btn-primary">
              Try again
            </button>
          )}
          <Link href={homePath} className="btn btn-quiet">
            Go home
          </Link>
        </div>
      </div>
    </div>
  );
}
