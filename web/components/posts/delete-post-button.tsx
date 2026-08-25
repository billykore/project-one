"use client";

/* eslint-disable react-hooks/set-state-in-effect */

import React, { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { ApiError, handleApiResponse } from "@/lib/errors";
import { ConfirmDialog } from "@/components/ui/modal";

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
    // ponytail: read from cookie
    const username = document.cookie.split("; ").find(r => r.startsWith("username="))?.split("=")[1] || null;
    setCurrentUser(username);
    setIsLoaded(true);
  }, []);

  const triggerRef = useRef<HTMLButtonElement | null>(null);

  const handleDelete = async () => {
    setIsDeleting(true);
    setError(null);
    try {
      await handleApiResponse(await fetch(`/api/posts/${postId}`, { method: "DELETE" }));
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
      <button ref={triggerRef} onClick={() => setIsOpen(true)} className="btn btn-danger">
        Delete post
      </button>

      <ConfirmDialog
        open={isOpen}
        onClose={() => setIsOpen(false)}
        onConfirm={handleDelete}
        rubric="This cannot be undone"
        title="Delete this post?"
        description="The post and its comments come down for everyone. There is no way to bring them back."
        confirmLabel="Delete post"
        pendingLabel="Deleting"
        pending={isDeleting}
        error={error}
      />
    </>
  );
}
