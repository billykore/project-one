"use client";

import React, { useState } from "react";
import { Textarea } from "@/components/ui/textarea";

interface CommentItemProps {
  comment: {
    id: number;
    username: string;
    content: string;
    created_at: string;
  };
  currentUser: string | null;
  onEdit: (id: number, content: string) => Promise<void>;
  onDelete: (id: number) => Promise<void>;
}

export function CommentItem({ comment, currentUser, onEdit, onDelete }: CommentItemProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(comment.content);
  const [isSaving, setIsSaving] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const canManage = currentUser !== null && comment.username === currentUser;

  const handleSave = async () => {
    if (!editContent.trim()) return;
    setIsSaving(true);
    setError(null);
    try {
      await onEdit(comment.id, editContent);
      setIsEditing(false);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to update comment";
      setError(message);
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async () => {
    if (confirm("Delete this comment?")) {
      setIsDeleting(true);
      setError(null);
      try {
        await onDelete(comment.id);
      } catch (err) {
        const message = err instanceof Error ? err.message : "Failed to delete comment";
        setError(message);
        setIsDeleting(false);
      }
    }
  };

  return (
    <div className="py-4">
      <div className="mb-2 flex items-baseline justify-between gap-3">
        <div className="flex items-baseline gap-2.5">
          <span className="meta font-medium text-ink">{comment.username}</span>
          <span className="meta">
            {new Date(comment.created_at).toLocaleDateString(undefined, {
              year: 'numeric',
              month: 'short',
              day: 'numeric',
              hour: '2-digit',
              minute: '2-digit'
            })}
          </span>
        </div>
        {canManage && !isEditing && (
          <div className="flex shrink-0 items-center gap-1">
            <button
              onClick={() => {
                setEditContent(comment.content);
                setIsEditing(true);
                setError(null);
              }}
              disabled={isDeleting}
              className="btn btn-ghost btn-sm"
            >
              Edit
            </button>
            <button
              onClick={handleDelete}
              disabled={isDeleting}
              className="btn btn-ghost btn-sm text-ember hover:text-ember"
            >
              {isDeleting ? "Deleting" : "Delete"}
            </button>
          </div>
        )}
      </div>

      {isEditing ? (
        <div className="mt-1 space-y-3">
          <Textarea
            value={editContent}
            onChange={(e) => setEditContent(e.target.value)}
            disabled={isSaving}
            rows={2}
          />
          {error && (
            <p className="font-mono text-xs text-ember" role="alert">
              {error}
            </p>
          )}
          <div className="flex justify-end gap-2">
            <button
              onClick={() => {
                setIsEditing(false);
                setEditContent(comment.content);
                setError(null);
              }}
              disabled={isSaving}
              className="btn btn-quiet btn-sm"
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              disabled={isSaving || !editContent.trim()}
              className="btn btn-primary btn-sm"
            >
              {isSaving ? "Saving" : "Save changes"}
            </button>
          </div>
        </div>
      ) : (
        <div>
          <p className="prose-body text-base whitespace-pre-wrap">{comment.content}</p>
          {error && (
            <p className="mt-1 font-mono text-xs text-ember" role="alert">
              {error}
            </p>
          )}
        </div>
      )}
    </div>
  );
}
