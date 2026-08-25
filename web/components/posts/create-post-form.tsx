"use client";

import React, { useActionState } from "react";
import Link from "next/link";
import { createPostAction } from "@/app/posts/create/actions";
import { FormState } from "@/lib/types/create-post.types";
import { InputField } from "@/components/ui/input";
import { TextAreaField } from "@/components/ui/textarea";
import { Spinner } from "@/components/ui/modal";

const initialState: FormState = {
  message: null,
  errors: {},
};

export function CreatePostForm() {
  const [state, formAction, isPending] = useActionState(createPostAction, initialState);

  return (
    <div className="panel w-full max-w-2xl p-6 sm:p-9">
      <header className="border-b border-rule pb-6">
        <p className="meta">New entry</p>
        <h1 className="mt-2 text-3xl font-semibold text-ink">Write a post</h1>
        <p className="prose-lead mt-2">
          A title, the piece itself, and a few tags so people can find it.
        </p>
      </header>

      <form action={formAction} className="mt-7 space-y-6" noValidate>
        <InputField
          label="Title"
          id="title"
          type="text"
          placeholder="What are you calling this?"
          error={state.errors?.title}
        />

        <TextAreaField
          label="Post"
          id="content"
          placeholder="Write at least a couple of sentences."
          error={state.errors?.content}
          rows={10}
        />

        <div>
          <InputField
            label="Tags"
            id="tags"
            type="text"
            placeholder="systems, cities, reading"
            error={state.errors?.tags}
          />
          <p className="meta mt-1.5 normal-case">Separate tags with commas</p>
        </div>

        {state.message && (
          <p className="notice notice-bad enter" role="alert">
            {state.message}
          </p>
        )}

        <div className="flex items-center justify-end gap-2 border-t border-rule pt-6">
          <Link href="/" className="btn btn-quiet btn-lg">
            Cancel
          </Link>
          <button type="submit" disabled={isPending} className="btn btn-primary btn-lg">
            {isPending && <Spinner />}
            {isPending ? "Publishing" : "Publish post"}
          </button>
        </div>
      </form>
    </div>
  );
}
