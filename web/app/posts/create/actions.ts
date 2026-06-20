"use server";

import { redirect } from "next/navigation";
import { revalidatePath } from "next/cache";
import { apiServer } from "@/lib/api-server";
import { ApiError } from "@/lib/api";
import { createPostSchema, FormState, CreatePostResponse } from "@/lib/types/create-post.types";

export async function createPostAction(
  prevState: FormState,
  formData: FormData
): Promise<FormState> {
  console.log("createPostAction FormData Entries:", Array.from(formData.entries()));
  const validatedFields = createPostSchema.safeParse({
    title: formData.get("title"),
    content: formData.get("content"),
    tags: formData.get("tags"),
  });
  console.log("createPostAction validation:", validatedFields);

  if (!validatedFields.success) {
    return {
      message: "Missing Fields. Failed to Create Post.",
    };
  }

  const { title, content, tags } = validatedFields.data;

  // Transform tags string to string array
  const tagArray = tags
    ? tags
        .split(",")
        .map((t: string) => t.trim())
        .filter((t: string) => t !== "")
    : [];

  try {
    await apiServer.post<CreatePostResponse>("/api/v1/posts", {
      title,
      content,
      tags: tagArray,
    });
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 401) {
        redirect("/login");
      }
      return {
        message: err.message || "Backend error. Failed to Create Post.",
      };
    }
    return {
      message: "Database Error: Failed to Create Post.",
    };
  }

  revalidatePath("/posts");
  redirect("/posts");
}
