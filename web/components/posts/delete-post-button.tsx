"use client";

/* eslint-disable react-hooks/set-state-in-effect */

import React, { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";

interface DeletePostButtonProps {
  postId: number;
  postAuthor?: string;
  redirectPath?: string;
  onSuccess?: () => void;
}

export default function DeletePostButton({
  postId,
  postAuthor,
  redirectPath = "/posts",
  onSuccess,
}: DeletePostButtonProps) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentUser, setCurrentUser] = useState<string | null>(null);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    const username = localStorage.getItem("username");
    setCurrentUser(username);
    setIsLoaded(true);
  }, []);

  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const cancelBtnRef = useRef<HTMLButtonElement | null>(null);

  const wasOpenRef = useRef(false);

  // Manage focus when modal opens/closes
  useEffect(() => {
    if (isOpen) {
      // Focus cancel button for safety and accessibility
      cancelBtnRef.current?.focus();
      wasOpenRef.current = true;
    } else if (wasOpenRef.current) {
      // Return focus to trigger button only if the modal was previously open
      triggerRef.current?.focus();
      wasOpenRef.current = false;
    }
  }, [isOpen]);

  // Close modal on Escape key press (only register when open)
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        setIsOpen(false);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen]);

  const handleDelete = async () => {
    setIsDeleting(true);
    setError(null);
    try {
      await api.delete(`/api/v1/posts/${postId}`);
      if (onSuccess) {
        onSuccess();
      } else {
        if (redirectPath) {
          router.push(redirectPath);
        }
        router.refresh();
      }
    } catch (err) {
      console.error("Failed to delete post", err);
      const message = err instanceof ApiError ? err.message : "Failed to delete post";
      setError(message);
      setIsDeleting(false);
    }
  };

  if (postAuthor) {
    if (!isLoaded) return null;
    if (currentUser !== postAuthor) return null;
  }

  return (
    <>
      <button
        ref={triggerRef}
        onClick={() => setIsOpen(true)}
        className="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 transition-colors cursor-pointer"
      >
        Delete Post
      </button>

      {isOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
          onClick={() => setIsOpen(false)}
        >
          <div
            role="dialog"
            aria-modal="true"
            aria-labelledby="delete-modal-title"
            aria-describedby="delete-modal-desc"
            className="w-full max-w-sm rounded-lg bg-white p-6 shadow-xl dark:bg-zinc-900"
            onClick={(e) => e.stopPropagation()}
          >
            <h3
              id="delete-modal-title"
              className="text-xl font-bold text-gray-900 dark:text-zinc-50"
            >
              Confirm Delete
            </h3>
            <p
              id="delete-modal-desc"
              className="mt-4 text-gray-600 dark:text-zinc-400"
            >
              Are you sure you want to delete this post? This action cannot be undone.
            </p>

            {error && (
              <p className="mt-2 text-sm font-medium text-red-600 dark:text-red-400">
                Error: {error}
              </p>
            )}

            <div className="mt-8 flex justify-end gap-3">
              <button
                ref={cancelBtnRef}
                onClick={() => setIsOpen(false)}
                disabled={isDeleting}
                className="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none disabled:opacity-50 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-300 dark:hover:bg-zinc-700 cursor-pointer"
              >
                Cancel
              </button>
              <button
                onClick={handleDelete}
                disabled={isDeleting}
                className="inline-flex items-center rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 disabled:opacity-50 cursor-pointer"
              >
                {isDeleting ? (
                  <>
                    <svg
                      className="-ml-1 mr-2 h-4 w-4 animate-spin text-white"
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
                      ></circle>
                      <path
                        className="opacity-75"
                        fill="currentColor"
                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                      ></path>
                    </svg>
                    Deleting...
                  </>
                ) : (
                  "Delete"
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
