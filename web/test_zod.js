import { z } from 'zod';

const createPostSchema = z.object({
  title: z.string().trim().min(1, "Title is required"),
  content: z.string().trim().min(1, "Content is required").min(10, "Content must be at least 10 characters long"),
  tags: z.string().optional(),
});

const res = createPostSchema.safeParse({
  title: "",
  content: "Lorem ipsum dolor sit amet",
  tags: "tag1"
});

console.log("SafeParse Result:", res);
console.log("JSON Stringified:", JSON.stringify(res));
