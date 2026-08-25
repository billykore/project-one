import React from "react";

type AvatarSize = "sm" | "md" | "lg";

const SIZES: Record<AvatarSize, string> = {
  sm: "h-7 w-7 text-[0.65rem]",
  md: "h-9 w-9 text-xs",
  lg: "h-16 w-16 text-lg",
};

interface AvatarProps {
  /** Display name or username — the first character becomes the mark. */
  name: string;
  size?: AvatarSize;
  className?: string;
}

/**
 * An initial set in the margin voice inside a squared-off, accent-tinted
 * block — a filing mark rather than a photo slot.
 */
export function Avatar({ name, size = "md", className = "" }: AvatarProps) {
  return (
    <span aria-hidden="true" className={`mark ${SIZES[size]} ${className}`.trim()}>
      {name.charAt(0)}
    </span>
  );
}
