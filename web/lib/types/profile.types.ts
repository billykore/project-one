import { z } from "zod";

export interface UserProfile {
  username: string;
  email: string;
  name: string;
}

export interface FollowerInfo {
  username: string;
  name: string;
  followed_at: string;
  is_mutual: boolean;
}

export interface PostInfo {
  id: number;
  title: string;
  content: string;
  created_at: string;
  tags?: string[];
}

export const changePasswordSchema = z.object({
  oldPassword: z.string().min(8, "Current password must be at least 8 characters"),
  newPassword: z.string().min(8, "New password must be at least 8 characters"),
  confirmPassword: z.string().min(8, "Confirmation password must be at least 8 characters"),
}).refine((data) => data.newPassword === data.confirmPassword, {
  message: "Passwords do not match",
  path: ["confirmPassword"],
});

export type ChangePasswordFormData = z.infer<typeof changePasswordSchema>;

export const editProfileSchema = z.object({
  first_name: z.string().trim().min(3, "First name must be at least 3 characters").max(100, "First name must be at most 100 characters"),
  last_name: z.string().trim().min(3, "Last name must be at least 3 characters").max(100, "Last name must be at most 100 characters"),
  username: z.string().trim().min(3, "Username must be at least 3 characters").max(30, "Username must be at most 30 characters").regex(/^[a-zA-Z0-9_]+$/, "Username may only contain letters, numbers, and underscores"),
});

export type EditProfileFormData = z.infer<typeof editProfileSchema>;
