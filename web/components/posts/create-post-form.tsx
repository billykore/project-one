"use client";

import React, { useActionState } from "react";
import Link from "next/link";
import { createPostAction } from "@/app/(post)/posts/create/actions";
import { FormState } from "@/lib/types/create-post.types";
import { InputField } from "@/components/ui/input";
import { TextAreaField } from "@/components/ui/textarea";

const initialState: FormState = {
  message: null,
  errors: {},
};

export function CreatePostForm() {
  const [state, formAction, isPending] = useActionState(createPostAction, initialState);

  return (
    <div className="max-w-2xl w-full space-y-8 bg-white p-10 rounded-xl shadow-lg border border-gray-100 dark:bg-zinc-900 dark:border-zinc-800">
      <div>
        <h2 className="text-center text-3xl font-extrabold text-gray-900 dark:text-zinc-50">
          Create a New Post
        </h2>
        <p className="mt-2 text-center text-sm text-gray-600 dark:text-zinc-400">
          Share your thoughts with the community
        </p>
      </div>
      <form action={formAction} className="mt-8 space-y-6" noValidate>
        <div className="space-y-4">
          <InputField
            label="Title"
            id="title"
            type="text"
            placeholder="Enter post title"
            error={state.errors?.title}
          />

          <TextAreaField
            label="Content"
            id="content"
            placeholder="What's on your mind? (min 10 characters)"
            error={state.errors?.content}
            rows={6}
          />

          <InputField
            label="Tags"
            id="tags"
            type="text"
            placeholder="e.g. news, technology, life (comma separated)"
            error={state.errors?.tags}
          />
        </div>

        {state.message && (
          <div className="rounded-md bg-red-50 p-4 border border-red-200 dark:bg-red-900/20 dark:border-red-800">
            <p className="text-sm text-red-700 dark:text-red-400">{state.message}</p>
          </div>
        )}

        <div className="flex items-center justify-between gap-4">
          <Link
            href="/"
            className="flex-1 text-center py-2 px-4 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 dark:bg-zinc-800 dark:border-zinc-700 dark:text-zinc-300 dark:hover:bg-zinc-700"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={isPending}
            className={`flex-1 flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white ${
              isPending
                ? "bg-indigo-400 cursor-not-allowed"
                : "bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            }`}
          >
            {isPending ? "Creating..." : "Create Post"}
          </button>
        </div>
      </form>
    </div>
  );
}
