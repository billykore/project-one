import { z } from "zod";

export const loginSchema = z.object({
  email: z.string().min(1, "Email is required").pipe(z.email("Please enter a valid email address")),
  password: z
    .string()
    .min(1, "Password is required")
    .min(8, "Password must be at least 8 characters long"),
});

export type LoginFormData = z.infer<typeof loginSchema>;

export interface LoginErrors {
  email?: string;
  password?: string;
  general?: string;
}
