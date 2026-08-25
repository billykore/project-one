"use client";

import React from "react";
import { Comment } from "@/lib/types/post.types";
import { CommentItem } from "./comment-item";

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
      <p className="meta py-8 text-center">No comments yet &mdash; say the first thing</p>
    );
  }

  return (
    <div className="divide-y divide-rule border-t border-rule">
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
