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
      <div className="panel p-6 sm:p-8">
        <p className="meta">Your details</p>
        <h1 className="mt-2 text-2xl font-semibold text-ink">Edit profile</h1>
        <p className="prose-lead mt-2 mb-6">
          Your name and username are shown on everything you post.
        </p>

        <form onSubmit={handleSubmit} className="space-y-4" noValidate>
          <InputField
            label="First name"
            id="first_name"
            type="text"
            value={form.first_name}
            onChange={handleChange}
            error={errors.first_name}
            placeholder="Ada"
            autoComplete="given-name"
          />

          <InputField
            label="Last name"
            id="last_name"
            type="text"
            value={form.last_name}
            onChange={handleChange}
            error={errors.last_name}
            placeholder="Lovelace"
            autoComplete="family-name"
          />

          <InputField
            label="Username"
            id="username"
            type="text"
            value={form.username}
            onChange={handleChange}
            error={errors.username}
            placeholder="ada"
            autoComplete="username"
          />

          {serverError && (
            <p className="notice notice-bad" role="alert">
              {serverError}
            </p>
          )}

          <div className="flex justify-end gap-2 border-t border-rule pt-5">
            <Link href={`/${initialUsername}`} className="btn btn-quiet btn-lg">
              Cancel
            </Link>
            <button type="submit" disabled={isSubmitting} className="btn btn-primary btn-lg">
              {isSubmitting ? "Saving" : "Save changes"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
