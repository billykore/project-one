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
