"use client";

import { useState, useEffect } from "react";
import { LikeButton } from "./like-button";
import { CommentForm } from "./comment-form";
import { CommentList } from "./comment-list";
import DeletePostButton from "./delete-post-button";
import { handleApiResponse } from "@/lib/errors";
import { Comment } from "@/lib/types/post.types";

interface PostInteractionSectionProps {
  postId: number;
  postAuthor?: string;
  initialComments?: Comment[];
  isGuest?: boolean;
  initialLikeCount?: number;
}

export default function PostInteractionSection({
  postId,
  postAuthor,
  initialComments = [],
  isGuest = false,
  initialLikeCount = 0,
}: PostInteractionSectionProps) {
  const [comments, setComments] = useState<Comment[]>(initialComments);
  const [isLiked, setIsLiked] = useState(false);
  const [likeCount, setLikeCount] = useState(initialLikeCount);
  const [loadingLike, setLoadingLike] = useState(true);
  const [currentUser, setCurrentUser] = useState<string | null>(null);

  useEffect(() => {
    async function init() {
      // ponytail: read from cookie
      const stored = isGuest ? null : document.cookie.split("; ").find(r => r.startsWith("username="))?.split("=")[1] || null;
      setCurrentUser(stored);

      if (isGuest) {
        setIsLiked(false);
        setLikeCount(initialLikeCount);
        setLoadingLike(false);
        return;
      }

      try {
        const res = await handleApiResponse<{ liked: boolean; like_count: number }>(
          await fetch(`/api/posts/${postId}/likes`)
        );
        setIsLiked(res.liked);
        setLikeCount(res.like_count);
      } catch (e) {
        console.error("Failed to load like status", e);
      } finally {
        setLoadingLike(false);
      }
    }
    init();
  }, [postId, isGuest, initialLikeCount]);

  const handleLikeToggle = async () => {
    if (isGuest) return;
    const originalIsLiked = isLiked;
    const originalLikeCount = likeCount;

    // Optimistic Update
    setIsLiked(!originalIsLiked);
    setLikeCount(originalIsLiked ? originalLikeCount - 1 : originalLikeCount + 1);

    try {
      if (originalIsLiked) {
        await handleApiResponse(await fetch(`/api/posts/${postId}/likes`, { method: "DELETE" }));
      } else {
        await handleApiResponse(await fetch(`/api/posts/${postId}/likes`, { method: "POST" }));
      }
    } catch (err) {
      console.error("Failed to save like action", err);
      // Rollback on failure
      setIsLiked(originalIsLiked);
      setLikeCount(originalLikeCount);
    }
  };

  const fetchComments = async () => {
    try {
      const res = await handleApiResponse<{ comments?: Comment[] }>(await fetch(`/api/posts/${postId}`));
      if (res && res.comments) {
        setComments(res.comments);
      }
    } catch (e) {
      console.error("Failed to fetch comments", e);
    }
  };

  const handleAddComment = async (content: string) => {
    if (isGuest) return;
    await handleApiResponse(await fetch(`/api/posts/${postId}/comments`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ content, id: postId }),
    }));
    await fetchComments();
  };

  const handleEditComment = async (commentId: number, content: string) => {
    if (isGuest) return;
    await handleApiResponse(await fetch(`/api/comments/${commentId}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ content }),
    }));
    await fetchComments();
  };

  const handleDeleteComment = async (commentId: number) => {
    if (isGuest) return;
    await handleApiResponse(await fetch(`/api/comments/${commentId}`, { method: "DELETE" }));
    await fetchComments();
  };

  return (
    <div className="mt-8 border-t border-gray-200 pt-8 dark:border-zinc-800 space-y-6">
      <div className="flex items-center justify-between space-x-4">
        <LikeButton
          isLiked={isLiked}
          likeCount={likeCount}
          isLoading={loadingLike}
          onToggle={handleLikeToggle}
          isGuest={isGuest}
        />
        {!isGuest && (
          <DeletePostButton postId={postId} postAuthor={postAuthor} />
        )}
      </div>

      <div className="space-y-4">
        <h3 className="text-lg font-bold text-gray-900 dark:text-zinc-50">
          Comments ({comments.length})
        </h3>
        {currentUser ? (
          <CommentForm onSubmit={handleAddComment} />
        ) : (
          <p className="text-sm text-gray-500">Please log in to add comments.</p>
        )}
        <CommentList
          comments={comments}
          currentUser={currentUser}
          onEditComment={handleEditComment}
          onDeleteComment={handleDeleteComment}
        />
      </div>
    </div>
  );
}
