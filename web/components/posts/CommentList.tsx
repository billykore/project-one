"use client";

import React from "react";
import { Comment } from "@/app/posts/model";
import { CommentItem } from "./CommentItem";

interface CommentListProps {
  comments: Comment[];
  currentUser: string | null;
  onEditComment: (id: number, content: string) => Promise<void>;
  onDeleteComment: (id: number) => Promise<void>;
}

export function CommentList({
  comments,
  currentUser,
  onEditComment,
  onDeleteComment,
}: CommentListProps) {
  if (comments.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-zinc-400">
        <p className="text-sm">No comments yet. Be the first to share your thoughts!</p>
      </div>
    );
  }

  return (
    <div className="space-y-1">
      {comments.map((comment) => (
        <CommentItem
          key={comment.id}
          comment={comment}
          currentUser={currentUser}
          onEdit={onEditComment}
          onDelete={onDeleteComment}
        />
      ))}
    </div>
  );
}
