"use client";

import React, { useState } from "react";
import { Textarea } from "@/components/ui/textarea";
import { Spinner } from "@/components/ui/modal";

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
          placeholder="Add to the conversation"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          disabled={isSubmitting}
          rows={3}
          required
        />
      </div>
      {error && (
        <p className="font-mono text-xs text-ember" role="alert">
          {error}
        </p>
      )}
      <div className="flex justify-end">
        <button
          type="submit"
          disabled={isSubmitting || !content.trim()}
          className="btn btn-primary"
        >
          {isSubmitting && <Spinner />}
          {isSubmitting ? "Posting" : "Post comment"}
        </button>
      </div>
    </form>
  );
}
