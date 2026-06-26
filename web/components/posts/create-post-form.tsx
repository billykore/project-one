"use client";

import React, { useActionState } from "react";
import Link from "next/link";
import { createPostAction } from "@/app/posts/create/actions";
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
    <div className="w-full max-w-2xl space-y-8 rounded-2xl bg-white p-8 shadow-sm ring-1 ring-gray-200 dark:bg-gray-900 dark:ring-gray-800 sm:p-10">
      <div className="space-y-1">
        <h2 className="text-2xl/[1.2] font-semibold tracking-tight text-gray-900 dark:text-white">
          Create a New Post
        </h2>
        <p className="text-sm text-gray-500 dark:text-gray-400 font-light">
          Share your thoughts with the community.
        </p>
      </div>
      <form action={formAction} className="space-y-5" noValidate>
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

        {state.message && (
          <div className="animate-in fade-in rounded-lg border border-red-200 bg-red-50 p-3 dark:border-red-800 dark:bg-red-950">
            <p className="text-sm text-red-600 dark:text-red-400">{state.message}</p>
          </div>
        )}

        <div className="flex items-center gap-3 pt-1">
          <Link
            href="/"
            className="flex-1 rounded-lg border border-gray-200 bg-white px-4 py-2.5 text-center text-sm font-medium text-gray-700 shadow-sm transition-all hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={isPending}
            className={`flex-1 inline-flex items-center justify-center rounded-lg px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 dark:focus:ring-offset-gray-950
              ${
                isPending
                  ? "cursor-not-allowed bg-indigo-400"
                  : "bg-linear-to-r from-indigo-600 to-indigo-700 hover:from-indigo-500 hover:to-indigo-600 hover:shadow-md hover:shadow-indigo-500/20 active:scale-[0.98]"
              }`}
          >
            {isPending && (
              <svg className="mr-2 h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
            )}
            {isPending ? "Creating…" : "Create Post"}
          </button>
        </div>
      </form>
    </div>
  );
}
