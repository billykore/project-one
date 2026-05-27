"use client";

import { useState, useEffect } from "react";
import { LikeButton } from "./LikeButton";
import { CommentForm } from "./CommentForm";
import { CommentList } from "./CommentList";
import { api } from "@/lib/api";
import { Comment } from "@/app/posts/model";

interface PostInteractionSectionProps {
  postId: number;
  initialComments?: Comment[];
}

export default function PostInteractionSection({
  postId,
  initialComments = [],
}: PostInteractionSectionProps) {
  const [comments, setComments] = useState<Comment[]>(initialComments);
  const [isLiked, setIsLiked] = useState(false);
  const [likeCount, setLikeCount] = useState(0);
  const [loadingLike, setLoadingLike] = useState(true);
  const [currentUser, setCurrentUser] = useState<string | null>(null);

  useEffect(() => {
    async function init() {
      const stored = localStorage.getItem("username");
      setCurrentUser(stored);

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
  }, [postId]);

  const handleLikeToggle = async () => {
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

  const handleAddComment = async (content: string) => {
    const response = await api.post<Comment>(`/api/v1/posts/${postId}/comments`, {
      content,
      id: postId,
    });
    setComments((prev) => [...prev, response]);
  };

  const handleEditComment = async (commentId: number, content: string) => {
    await api.put(`/api/v1/comments/${commentId}`, { content });
    setComments((prev) =>
      prev.map((c) => (c.id === commentId ? { ...c, content } : c))
    );
  };

  const handleDeleteComment = async (commentId: number) => {
    await api.delete(`/api/v1/comments/${commentId}`);
    setComments((prev) => prev.filter((c) => c.id !== commentId));
  };

  return (
    <div className="mt-8 border-t border-gray-200 pt-8 dark:border-zinc-800 space-y-6">
      <div className="flex items-center gap-4">
        <LikeButton
          isLiked={isLiked}
          likeCount={likeCount}
          isLoading={loadingLike}
          onToggle={handleLikeToggle}
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
