import React from "react";

export interface LikeButtonProps {
  isLiked: boolean;
  likeCount: number;
  isLoading: boolean;
  onToggle: () => void;
  isGuest?: boolean;
}

export function LikeButton({
  isLiked,
  likeCount,
  isLoading,
  onToggle,
  isGuest = false,
}: LikeButtonProps) {
  const buttonContent = (
    <>
      {isLoading ? (
        <svg
          className="animate-spin h-5 w-5 text-gray-500 dark:text-zinc-400"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      ) : isLiked ? (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 24 24"
          fill="currentColor"
          className="h-5 w-5 text-red-500 transition-colors"
        >
          <path d="M11.645 20.91l-.007-.003-.022-.012a15.247 15.247 0 01-.383-.218 25.18 25.18 0 01-4.244-3.17C4.688 15.36 2.25 12.174 2.25 8.25 2.25 5.322 4.714 3 7.688 3A5.5 5.5 0 0112 5.052 5.5 5.5 0 0116.313 3c2.973 0 5.437 2.322 5.437 5.25 0 3.925-2.438 7.111-4.739 9.256a25.175 25.175 0 01-4.244 3.17 15.247 15.247 0 01-.383.219l-.022.012-.007.004-.003.001a.752.752 0 01-.704 0l-.003-.001z" />
        </svg>
      ) : (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          strokeWidth={1.5}
          stroke="currentColor"
          className="h-5 w-5 text-gray-500 hover:text-red-500 dark:text-zinc-400 dark:hover:text-red-400 transition-colors"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12Z"
          />
        </svg>
      )}
      <span className={isLiked ? "text-red-600 dark:text-red-400 font-semibold" : ""}>
        {likeCount}
      </span>
    </>
  );

  if (isGuest) {
    return (
      <span title="Please log in to like this post" className="inline-block cursor-not-allowed">
        <button
          disabled
          aria-label={isLiked ? "Unlike post" : "Like post"}
          className="inline-flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-700 pointer-events-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 select-none opacity-50"
        >
          {buttonContent}
        </button>
      </span>
    );
  }

  return (
    <button
      onClick={onToggle}
      disabled={isLoading}
      aria-label={isLiked ? "Unlike post" : "Like post"}
      className="inline-flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 active:scale-95 transition-all duration-200 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed select-none cursor-pointer"
    >
      {buttonContent}
    </button>
  );
}
