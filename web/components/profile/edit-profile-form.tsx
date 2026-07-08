"use client";

import React, { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { InputField } from "@/components/ui/input";
import { editProfileSchema, EditProfileFormData } from "@/lib/types/profile.types";
import { handleApiResponse, ApiError } from "@/lib/errors";

interface EditProfileFormProps {
  initialFirstName: string;
  initialLastName: string;
  initialUsername: string;
}

export default function EditProfileForm({
  initialFirstName,
  initialLastName,
  initialUsername,
}: EditProfileFormProps) {
  const router = useRouter();

  const [form, setForm] = useState<EditProfileFormData>({
    first_name: initialFirstName,
    last_name: initialLastName,
    username: initialUsername,
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [serverError, setServerError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setForm((prev) => ({ ...prev, [name]: value }));
    // Clear field-level error on edit
    if (errors[name]) {
      setErrors((prev) => ({ ...prev, [name]: "" }));
    }
    if (serverError) {
      setServerError(null);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setServerError(null);
    setErrors({});

    // Client-side Zod validation
    const result = editProfileSchema.safeParse(form);
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      result.error.issues.forEach((issue) => {
        const path = issue.path[0] as string;
        fieldErrors[path] = issue.message;
      });
      setErrors(fieldErrors);
      return;
    }

    setIsSubmitting(true);
    try {
      await handleApiResponse(
        await fetch("/api/users/profile", {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(result.data),
        })
      );
      // Success — redirect to profile page with success param
      router.push(`/${result.data.username}?updated=1`);
    } catch (err) {
      if (err instanceof ApiError) {
        const msg = err.message.toLowerCase();
        if (msg.includes("already taken")) {
          setServerError("This username is already taken. Please choose another.");
        } else {
          setServerError(err.message);
        }
      } else {
        setServerError("Unable to connect. Please try again.");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="mx-auto w-full max-w-lg">
      <div className="rounded-2xl border border-zinc-200 bg-white/80 p-6 shadow-sm backdrop-blur dark:border-zinc-800 dark:bg-zinc-900/80">
        <h2 className="text-xl font-bold text-gray-900 dark:text-zinc-50 mb-1">Edit Profile</h2>
        <p className="text-sm text-zinc-500 dark:text-zinc-400 mb-6">
          Update your first name, last name, or username.
        </p>

        <form onSubmit={handleSubmit} className="space-y-4" noValidate>
          <InputField
            label="First Name"
            id="first_name"
            type="text"
            value={form.first_name}
            onChange={handleChange}
            error={errors.first_name}
            placeholder="Enter your first name"
            autoComplete="given-name"
          />

          <InputField
            label="Last Name"
            id="last_name"
            type="text"
            value={form.last_name}
            onChange={handleChange}
            error={errors.last_name}
            placeholder="Enter your last name"
            autoComplete="family-name"
          />

          <InputField
            label="Username"
            id="username"
            type="text"
            value={form.username}
            onChange={handleChange}
            error={errors.username}
            placeholder="Enter your username"
            autoComplete="username"
          />

          {serverError && (
            <div className="rounded-md bg-red-50 p-3 text-sm text-red-700 dark:bg-red-900/10 dark:text-red-400" role="alert">
              {serverError}
            </div>
          )}

          <div className="flex gap-3 pt-2">
            <button
              type="submit"
              disabled={isSubmitting}
              className="flex-1 rounded-lg bg-indigo-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition duration-200 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:opacity-50 cursor-pointer dark:bg-indigo-500 dark:hover:bg-indigo-600"
            >
              {isSubmitting ? "Saving..." : "Save Changes"}
            </button>
            <Link
              href={`/${initialUsername}`}
              className="flex-1 rounded-lg border border-zinc-200 bg-white px-4 py-2.5 text-sm font-semibold text-zinc-700 text-center shadow-sm transition duration-200 hover:bg-zinc-50 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-300 dark:hover:bg-zinc-700 focus:outline-none focus:ring-2 focus:ring-zinc-400 focus:ring-offset-2"
            >
              Cancel
            </Link>
          </div>
        </form>
      </div>
    </div>
  );
}
