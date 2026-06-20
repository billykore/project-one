"use client";

import { useState, useEffect } from "react";
import { LikeButton } from "./LikeButton";
import { CommentForm } from "./CommentForm";
import { CommentList } from "./CommentList";
import { api } from "@/lib/api";
import { Comment } from "@/lib/types/post.types";

interface PostInteractionSectionProps {
  postId: number;
  initialComments?: Comment[];
  isGuest?: boolean;
  initialLikeCount?: number;
}

export default function PostInteractionSection({
  postId,
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
      const stored = isGuest ? null : localStorage.getItem("username");
      setCurrentUser(stored);

      if (isGuest) {
        setIsLiked(false);
        setLikeCount(initialLikeCount);
        setLoadingLike(false);
        return;
      }

      try {
        const res = await api.get<{ liked: boolean; like_count: number }>(
          `/api/v1/posts/${postId}/likes`
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
        await api.delete(`/api/v1/posts/${postId}/likes`);
      } else {
        await api.post(`/api/v1/posts/${postId}/likes`, {});
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
      const res = await api.get<{ comments?: Comment[] }>(`/api/v1/posts/${postId}`);
      if (res && res.comments) {
        setComments(res.comments);
      }
    } catch (e) {
      console.error("Failed to fetch comments", e);
    }
  };

  const handleAddComment = async (content: string) => {
    if (isGuest) return;
    await api.post(`/api/v1/posts/${postId}/comments`, {
      content,
      id: postId,
    });
    await fetchComments();
  };

  const handleEditComment = async (commentId: number, content: string) => {
    if (isGuest) return;
    await api.put(`/api/v1/comments/${commentId}`, { content });
    await fetchComments();
  };

  const handleDeleteComment = async (commentId: number) => {
    if (isGuest) return;
    await api.delete(`/api/v1/comments/${commentId}`);
    await fetchComments();
  };

  return (
    <div className="mt-8 border-t border-gray-200 pt-8 dark:border-zinc-800 space-y-6">
      <div className="flex items-center gap-4">
        <LikeButton
          isLiked={isLiked}
          likeCount={likeCount}
          isLoading={loadingLike}
          onToggle={handleLikeToggle}
          isGuest={isGuest}
        />
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
