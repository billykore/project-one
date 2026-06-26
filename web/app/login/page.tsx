"use client";

import React, { useState, use } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ZodError } from "zod";
import { ApiError } from "@/lib/errors";
import { InputField } from "@/components/ui/input";
import { useErrorModal } from "@/hooks/use-error-modal";
import { loginRequestBodySchema, LoginRequestBody } from "@/app/api/login/schema";

interface LoginErrors {
  email?: string;
  password?: string;
  general?: string;
}

export default function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ registered?: string }>;
}) {
  const { registered } = use(searchParams);
  const router = useRouter();
  const { showError } = useErrorModal();
  const [formData, setFormData] = useState<LoginRequestBody>({
    email: "",
    password: "",
  });
  const [errors, setErrors] = useState<LoginErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    if (errors[name as keyof LoginRequestBody]) {
      setErrors((prev) => ({ ...prev, [name]: undefined }));
    }
  };

  const validate = (): boolean => {
    try {
      loginRequestBodySchema.parse(formData);
      setErrors({});
      return true;
    } catch (err) {
      if (err instanceof ZodError) {
        const fieldErrors: LoginErrors = {};
        err.issues.forEach((e) => {
          if (e.path.length > 0) {
            const path = e.path[0] as keyof LoginRequestBody;
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

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!validate()) return;

    setIsSubmitting(true);
    setErrors({});

    try {
      const res = await fetch("/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email: formData.email.trim(),
          password: formData.password,
        }),
      });

      if (!res.ok) {
        const errBody = await res.json().catch(() => ({ error: "Login failed" }));
        throw new ApiError(errBody.error || "Login failed", res.status);
      }

      router.push("/");
    } catch (err) {
      if (err instanceof ApiError && (err.status === 401 || err.status === 400)) {
        setErrors({ general: err.message });
        return;
      }
      const errorMessage = err instanceof Error ? err.message : "An error occurred. Please try again later.";
      showError(errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8 bg-white p-10 rounded-xl shadow-lg">
        {registered === "true" && (
          <div className="rounded-md bg-green-50 border border-green-200 p-4">
            <p className="text-sm text-green-800">
              ✅ Registration successful! Please sign in with your new account.
            </p>
          </div>
        )}
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Sign in to your account
          </h2>
        </div>
        <form className="mt-8 space-y-6" onSubmit={handleSubmit} noValidate>
          <div className="rounded-md space-y-4">
            <InputField
              label="Email address"
              id="email"
              type="email"
              autoComplete="email"
              placeholder="Email address"
              value={formData.email}
              onChange={handleChange}
              error={errors.email}
            />
            <InputField
              label="Password"
              id="password"
              type="password"
              autoComplete="current-password"
              placeholder="Password"
              value={formData.password}
              onChange={handleChange}
              error={errors.password}
            />
          </div>

          <div className="flex items-center justify-between">
            <div className="text-sm">
              <Link
                href="/forgot-password"
                title="Forgot password link"
                className="font-medium text-indigo-600 hover:text-indigo-500"
              >
                Forgot your password?
              </Link>
            </div>
            <div className="text-sm">
              <Link
                href="/register"
                title="Create an account"
                className="font-medium text-indigo-600 hover:text-indigo-500"
              >
                Create an account
              </Link>
            </div>
          </div>

          {errors.general && (
            <div className="rounded-md bg-red-50 p-4">
              <p className="text-sm text-red-700">{errors.general}</p>
            </div>
          )}

          <div>
            <button
              type="submit"
              disabled={isSubmitting}
              className={`group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white ${
                isSubmitting
                  ? "bg-indigo-400 cursor-not-allowed"
                  : "bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              }`}
            >
              {isSubmitting ? "Signing in..." : "Sign in"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
