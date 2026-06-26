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
    <div className="border-b border-gray-100 py-4 last:border-0 dark:border-gray-800">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <span className="font-semibold text-sm text-gray-900 dark:text-white">
            {comment.username}
          </span>
          <span className="text-xs text-gray-500">
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
          <div className="flex items-center gap-2 text-xs">
            <button
              onClick={() => {
                setEditContent(comment.content);
                setIsEditing(true);
                setError(null);
              }}
              disabled={isDeleting}
              className="text-indigo-600 hover:text-indigo-800 font-medium disabled:opacity-50 dark:text-indigo-400 dark:hover:text-indigo-300"
            >
              Edit
            </button>
            <span className="text-gray-300 dark:text-zinc-700">|</span>
            <button
              onClick={handleDelete}
              disabled={isDeleting}
              className="text-red-600 hover:text-red-800 font-medium disabled:opacity-50 dark:text-red-400 dark:hover:text-red-300"
            >
              {isDeleting ? "Deleting..." : "Delete"}
            </button>
          </div>
        )}
      </div>

      {isEditing ? (
        <div className="space-y-3 mt-1">
          <Textarea
            value={editContent}
            onChange={(e) => setEditContent(e.target.value)}
            disabled={isSaving}
            rows={2}
          />
          {error && (
            <p className="text-xs text-red-600 dark:text-red-400" role="alert">
              {error}
            </p>
          )}
          <div className="flex gap-2 justify-end">
            <button
              onClick={() => {
                setIsEditing(false);
                setEditContent(comment.content);
                setError(null);
              }}
              disabled={isSaving}
              className="px-3 py-1.5 text-xs font-medium rounded border border-gray-300 text-gray-700 hover:bg-gray-50 dark:border-zinc-700 dark:text-zinc-300 dark:hover:bg-zinc-800"
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              disabled={isSaving || !editContent.trim()}
              className="px-3 py-1.5 text-xs font-medium rounded text-white bg-indigo-600 hover:bg-indigo-700 dark:bg-indigo-500 dark:hover:bg-indigo-600 disabled:opacity-50"
            >
              {isSaving ? "Saving..." : "Save"}
            </button>
          </div>
        </div>
      ) : (
        <div>
          <p className="text-sm text-gray-700 dark:text-zinc-300 whitespace-pre-wrap">
            {comment.content}
          </p>
          {error && (
            <p className="text-xs text-red-600 mt-1 dark:text-red-400" role="alert">
              {error}
            </p>
          )}
        </div>
      )}
    </div>
  );
}
