"use client";

import React from "react";
import Link from "next/link";
import { useCreatePost } from "./controller";
import { InputField } from "@/components/ui/input";
import { TextAreaField } from "@/components/ui/textarea";

export function CreatePostForm() {
  const { formData, errors, isSubmitting, isLoading, handleChange, handleSubmit } = useCreatePost();

  if (isLoading) {
    return (
      <div className="max-w-2xl w-full flex items-center justify-center p-12 bg-white rounded-xl shadow-lg border border-gray-100">
        <div className="text-xl font-semibold text-gray-700">Checking authorization...</div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl w-full space-y-8 bg-white p-10 rounded-xl shadow-lg border border-gray-100">
      <div>
        <h2 className="text-center text-3xl font-extrabold text-gray-900">
          Create a New Post
        </h2>
        <p className="mt-2 text-center text-sm text-gray-600">
          Share your thoughts with the community
        </p>
      </div>
      <form className="mt-8 space-y-6" onSubmit={handleSubmit} noValidate>
        <div className="space-y-4">
          <InputField
            label="Title"
            id="title"
            type="text"
            placeholder="Enter post title"
            value={formData.title}
            onChange={handleChange}
            error={errors.title}
          />

          <TextAreaField
            label="Content"
            id="content"
            placeholder="What's on your mind? (min 10 characters)"
            value={formData.content}
            onChange={handleChange}
            error={errors.content}
            rows={6}
          />

          <InputField
            label="Tags"
            id="tags"
            type="text"
            placeholder="e.g. news, technology, life (comma separated)"
            value={formData.tags || ""}
            onChange={handleChange}
            error={errors.tags}
          />
        </div>

        {errors.general && (
          <div className="rounded-md bg-red-50 p-4 border border-red-200">
            <p className="text-sm text-red-700">{errors.general}</p>
          </div>
        )}

        <div className="flex items-center justify-between gap-4">
          <Link
            href="/home"
            className="flex-1 text-center py-2 px-4 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={isSubmitting}
            className={`flex-1 flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white ${
              isSubmitting
                ? "bg-indigo-400 cursor-not-allowed"
                : "bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            }`}
          >
            {isSubmitting ? "Creating..." : "Create Post"}
          </button>
        </div>
      </form>
    </div>
  );
}
