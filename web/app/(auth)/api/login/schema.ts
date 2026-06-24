import { z } from "zod";

export const loginRequestBodySchema = z.object({
  email: z.string().min(1, "Email is required").pipe(z.email("Please enter a valid email address")),
  password: z
    .string()
    .min(1, "Password is required")
    .min(8, "Password must be at least 8 characters long"),
});

export type LoginRequestBody = z.infer<typeof loginRequestBodySchema>;