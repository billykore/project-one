import { z } from "zod";

export const registerRequestBodySchema = z
  .object({
    first_name: z
      .string()
      .min(1, "First name is required")
      .min(3, "First name must be at least 3 characters long"),
    last_name: z
      .string()
      .min(1, "Last name is required")
      .min(3, "Last name must be at least 3 characters long"),
    username: z
      .string()
      .min(1, "Username is required")
      .min(3, "Username must be at least 3 characters long"),
    email: z
      .string()
      .min(1, "Email is required")
      .pipe(z.email("Please enter a valid email address")),
    password: z
      .string()
      .min(1, "Password is required")
      .min(8, "Password must be at least 8 characters long"),
    confirmPassword: z.string().min(1, "Please confirm your password"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  });

export type RegisterRequestBody = z.infer<typeof registerRequestBodySchema>;
