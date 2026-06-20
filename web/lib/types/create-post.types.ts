import { z } from "zod";

export const createPostSchema = z.object({
  title: z.string().trim().min(1, "Title is required"),
  content: z.string().trim().min(1, "Content is required").min(10, "Content must be at least 10 characters long"),
  tags: z.string().optional(),
});

export type CreatePostFormData = z.infer<typeof createPostSchema>;

export interface FormState {
  errors?: {
    title?: string[];
    content?: string[];
    tags?: string[];
    general?: string[];
  };
  message?: string | null;
}

export interface CreatePostResponse {
  id: number;
  message: string;
}
