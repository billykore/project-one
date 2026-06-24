"use client";

import React, { useEffect, useRef } from "react";
import Link from "next/link";
import { useErrorModal } from "@/hooks/use-error-modal";

export default function ErrorModal() {
  const { state, closeError } = useErrorModal();
  const { open, message, onRetry } = state;
  const primaryBtnRef = useRef<HTMLButtonElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && open) {
        closeError();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [open, closeError]);

  useEffect(() => {
    if (open) {
      previousFocusRef.current = document.activeElement as HTMLElement;
      primaryBtnRef.current?.focus();
    } else if (previousFocusRef.current) {
      previousFocusRef.current.focus();
    }
  }, [open]);

  if (!open) return null;

  const handleRetry = () => {
    if (onRetry) {
      onRetry();
    } else {
      window.location.reload();
    }
    closeError();
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm transition-opacity"
      onClick={closeError}
      role="dialog"
      aria-modal="true"
      aria-labelledby="error-modal-title"
      aria-describedby="error-modal-message"
    >
      <div
        className="w-full max-w-md rounded-xl bg-white p-8 shadow-xl dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 transform transition-all"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-6 flex justify-center">
          <svg
            className="h-12 w-12 text-red-500"
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

        <h2
          id="error-modal-title"
          className="text-center text-2xl font-bold text-zinc-900 dark:text-zinc-50 mb-2"
        >
          Something went wrong
        </h2>

        <p
          id="error-modal-message"
          className="text-center text-zinc-600 dark:text-zinc-400 mb-8"
        >
          {message}
        </p>

        <div className="flex flex-col gap-3">
          <button
            ref={primaryBtnRef}
            onClick={handleRetry}
            className="flex h-11 items-center justify-center rounded-full bg-zinc-900 px-8 text-sm font-medium text-white transition-colors hover:bg-zinc-700 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200 cursor-pointer"
          >
            Try again
          </button>
          <Link
            href="/"
            onClick={closeError}
            className="flex h-11 items-center justify-center rounded-full border border-zinc-200 px-8 text-sm font-medium text-zinc-900 transition-colors hover:bg-zinc-50 dark:border-zinc-800 dark:text-zinc-50 dark:hover:bg-zinc-900"
          >
            Go back home
          </Link>
        </div>
      </div>
    </div>
  );
}
