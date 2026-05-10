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
    <div className="flex flex-col items-center justify-center min-h-[60vh] p-4 text-center">
      <div className="max-w-md w-full bg-white dark:bg-zinc-900 p-8 rounded-lg shadow-sm border border-zinc-200 dark:border-zinc-800">
        <div className="mb-6">
          <svg
            className="mx-auto h-12 w-12 text-red-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
        </div>
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-50 mb-2">
          Something went wrong
        </h1>
        <p className="text-zinc-600 dark:text-zinc-400 mb-8">
          {message}
        </p>
        <div className="flex flex-col gap-3">
          <Link
            href={homePath}
            className="flex h-11 items-center justify-center rounded-full bg-zinc-900 px-8 text-sm font-medium text-white transition-colors hover:bg-zinc-700 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
          >
            Go back home
          </Link>
          {onRetry && (
            <button
              onClick={onRetry}
              className="flex h-11 items-center justify-center rounded-full border border-zinc-200 px-8 text-sm font-medium text-zinc-900 transition-colors hover:bg-zinc-50 dark:border-zinc-800 dark:text-zinc-50 dark:hover:bg-zinc-900"
            >
              Try again
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
