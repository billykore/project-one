import { z } from "zod";

export const createPostSchema = z.object({
  title: z.string().min(1, "Title is required"),
  content: z.string().min(1, "Content is required").min(10, "Content must be at least 10 characters long"),
  tags: z.string().optional(),
});

export type CreatePostFormData = z.infer<typeof createPostSchema>;

export interface CreatePostErrors {
  title?: string;
  content?: string;
  tags?: string;
  general?: string;
}

export interface CreatePostResponse {
  id: number;
  message: string;
}
