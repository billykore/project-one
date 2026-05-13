"use server";

import { redirect } from "next/navigation";
import { revalidatePath } from "next/cache";
import { apiServer } from "@/lib/api-server";
import { ApiError } from "@/lib/api";
import { createPostSchema, FormState, CreatePostResponse } from "./model";

export async function createPostAction(
  prevState: FormState,
  formData: FormData
): Promise<FormState> {
  const validatedFields = createPostSchema.safeParse({
    title: formData.get("title"),
    content: formData.get("content"),
    tags: formData.get("tags"),
  });

  if (!validatedFields.success) {
    return {
      errors: validatedFields.error.flatten().fieldErrors,
      message: "Missing Fields. Failed to Create Post.",
    };
  }

  const { title, content, tags } = validatedFields.data;

  // Transform tags string to string array
  const tagArray = tags
    ? tags
        .split(",")
        .map((t) => t.trim())
        .filter((t) => t !== "")
    : [];

  try {
    await apiServer.post<CreatePostResponse>("/posts", {
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
