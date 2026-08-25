import React from "react";
import { Spinner } from "@/components/ui/modal";

export interface LikeButtonProps {
  isLiked: boolean;
  likeCount: number;
  isLoading: boolean;
  onToggle: () => void;
  isGuest?: boolean;
}

function Heart({ filled }: { filled: boolean }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill={filled ? "currentColor" : "none"}
      stroke="currentColor"
      strokeWidth={1.6}
      strokeLinecap="round"
      strokeLinejoin="round"
      className="h-4 w-4"
      aria-hidden="true"
    >
      <path d="M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12Z" />
    </svg>
  );
}

export function LikeButton({
  isLiked,
  likeCount,
  isLoading,
  onToggle,
  isGuest = false,
}: LikeButtonProps) {
  const content = (
    <>
      {isLoading ? <Spinner className="h-4 w-4" /> : <Heart filled={isLiked} />}
      <span className="font-mono text-xs tracking-wider tabular-nums">{likeCount}</span>
    </>
  );

  const tone = isLiked ? "btn btn-marked" : "btn btn-quiet";

  if (isGuest) {
    return (
      <span title="Log in to like this post" className="inline-block cursor-not-allowed">
        <button
          disabled
          aria-label={isLiked ? "Unlike post" : "Like post"}
          className={`${tone} pointer-events-none`}
        >
          {content}
        </button>
      </span>
    );
  }

  return (
    <button
      onClick={onToggle}
      disabled={isLoading}
      aria-pressed={isLiked}
      aria-label={isLiked ? "Unlike post" : "Like post"}
      className={tone}
    >
      {content}
    </button>
  );
}
