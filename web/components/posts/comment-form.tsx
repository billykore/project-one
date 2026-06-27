"use client";

import React, { useState } from "react";
import { Textarea } from "@/components/ui/textarea";

interface CommentFormProps {
  onSubmit: (content: string) => Promise<void>;
}

export function CommentForm({ onSubmit }: CommentFormProps) {
  const [content, setContent] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) return;

    setIsSubmitting(true);
    setError(null);
    try {
      await onSubmit(content);
      setContent("");
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to post comment";
      setError(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label htmlFor="comment-content" className="sr-only">
          Add a comment
        </label>
        <Textarea
          id="comment-content"
          placeholder="Add a comment..."
          value={content}
          onChange={(e) => setContent(e.target.value)}
          disabled={isSubmitting}
          rows={3}
          required
        />
      </div>
      {error && (
        <p className="text-sm text-red-600 dark:text-red-400" role="alert">
          {error}
        </p>
      )}
      <div className="flex justify-end">
        <button
          type="submit"
          disabled={isSubmitting || !content.trim()}
          className={`inline-flex items-center rounded-lg px-4 py-2 text-sm font-medium text-white shadow-sm transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 dark:focus:ring-offset-gray-950 ${
            isSubmitting || !content.trim()
              ? "cursor-not-allowed bg-indigo-400 opacity-70"
              : "bg-linear-to-r from-indigo-600 to-indigo-700 hover:from-indigo-500 hover:to-indigo-600 hover:shadow-md hover:shadow-indigo-500/20 active:scale-[0.98]"
          }`}
        >
          {isSubmitting ? "Posting..." : "Post Comment"}
        </button>
      </div>
    </form>
  );
}
