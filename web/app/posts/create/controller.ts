"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { ZodError } from "zod";
import { api, ApiError } from "@/lib/api";
import { createPostSchema, CreatePostFormData, CreatePostErrors, CreatePostResponse } from "./model";

export const useCreatePost = () => {
  const router = useRouter();
  const [formData, setFormData] = useState<CreatePostFormData>({
    title: "",
    content: "",
    tags: "",
  });
  const [errors, setErrors] = useState<CreatePostErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const checkAuth = async () => {
      try {
        await api.get("/users/me");
        setIsLoading(false);
      } catch (err) {
        if (err instanceof ApiError && err.status === 401) {
          router.push("/login");
          return;
        }
        if (err instanceof Error) {
          router.push("/error?message=" + encodeURIComponent(err.message));
          return;
        }
        setIsLoading(false);
      }
    };

    checkAuth();
  }, [router]);

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));

    // Clear error for the field when user starts typing
    if (errors[name as keyof CreatePostFormData]) {
      setErrors((prev) => ({ ...prev, [name]: undefined }));
    }
  };

  const validate = (): boolean => {
    try {
      createPostSchema.parse(formData);
      setErrors({});
      return true;
    } catch (err) {
      if (err instanceof ZodError) {
        const fieldErrors: CreatePostErrors = {};
        err.issues.forEach((e) => {
          if (e.path.length > 0) {
            const path = e.path[0] as keyof CreatePostFormData;
            if (!fieldErrors[path]) {
              fieldErrors[path] = e.message;
            }
          } else {
            fieldErrors.general = e.message;
          }
        });
        setErrors(fieldErrors);
      }
      return false;
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) return;

    setIsSubmitting(true);
    setErrors({});

    try {
      // Transform tags string to string array
      const tagArray = formData.tags
        ? formData.tags
            .split(",")
            .map((t) => t.trim())
            .filter((t) => t !== "")
        : [];

      await api.post<CreatePostResponse>("/posts", {
        title: formData.title.trim(),
        content: formData.content.trim(),
        tags: tagArray,
      });

      // Redirect to home page on success
      router.push("/home");
    } catch (err) {
      if (err instanceof ApiError && (err.status === 400 || err.status === 401 || err.status === 403)) {
        setErrors({ general: err.message });
        return;
      }

      const errorMessage =
        err instanceof Error
          ? err.message
          : "An error occurred while creating the post. Please try again.";
      router.push(`/error?message=${encodeURIComponent(errorMessage)}`);
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
    formData,
    errors,
    isSubmitting,
    isLoading,
    handleChange,
    handleSubmit,
  };
};
