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
    <div className="min-h-screen grid lg:grid-cols-2 font-sans">
      {/* Brand Panel */}
      <div className="relative hidden lg:flex flex-col justify-between bg-linear-to-br from-indigo-600 via-indigo-700 to-purple-800 p-12 overflow-hidden">
        {/* Background decoration */}
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,rgba(255,255,255,0.1)_0%,transparent_60%)]" />
        <div className="absolute top-0 right-0 w-96 h-96 bg-linear-to-bl from-purple-500/30 to-transparent rounded-full blur-3xl" />
        <div className="absolute bottom-0 left-0 w-80 h-80 bg-linear-to-tr from-indigo-400/20 to-transparent rounded-full blur-3xl" />

        {/* Logo & Brand */}
        <div className="relative z-10">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-white/20 backdrop-blur-sm shadow-inner ring-1 ring-white/30">
              <svg className="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6A2.25 2.25 0 0 1 6 3.75h2.25A2.25 2.25 0 0 1 10.5 6v2.25a2.25 2.25 0 0 1-2.25 2.25H6a2.25 2.25 0 0 1-2.25-2.25V6ZM3.75 15.75A2.25 2.25 0 0 1 6 13.5h2.25a2.25 2.25 0 0 1 2.25 2.25V18a2.25 2.25 0 0 1-2.25 2.25H6A2.25 2.25 0 0 1 3.75 18v-2.25ZM13.5 6a2.25 2.25 0 0 1 2.25-2.25H18A2.25 2.25 0 0 1 20.25 6v2.25A2.25 2.25 0 0 1 18 10.5h-2.25a2.25 2.25 0 0 1-2.25-2.25V6ZM13.5 15.75a2.25 2.25 0 0 1 2.25-2.25H18a2.25 2.25 0 0 1 2.25 2.25V18A2.25 2.25 0 0 1 18 20.25h-2.25A2.25 2.25 0 0 1 13.5 18v-2.25Z" />
              </svg>
            </div>
            <span className="text-xl font-semibold text-white tracking-tight">Project1</span>
          </div>
        </div>

        {/* Tagline */}
        <div className="relative z-10 space-y-4">
          <h1 className="text-4xl/[1.15] font-semibold text-white tracking-tight">
            Welcome back.
            <br />
            <span className="text-indigo-200">Sign in to continue.</span>
          </h1>
          <p className="text-indigo-200/80 text-sm/[1.6] font-light max-w-xs">
            Access your dashboard, connect with your community, and pick up
            right where you left off.
          </p>
        </div>

        {/* Footer */}
        <p className="relative z-10 text-xs font-light text-indigo-300/50 tracking-wide">
          &copy; {new Date().getFullYear()} Project1. All rights reserved.
        </p>
      </div>

      {/* Form Panel */}
      <div className="flex items-center justify-center px-4 py-12 sm:px-6 lg:px-12 bg-white dark:bg-gray-950">
        <div className="w-full max-w-md space-y-8">
          {/* Mobile logo */}
          <div className="lg:hidden flex flex-col items-center gap-2">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-linear-to-br from-indigo-600 to-purple-700 shadow-lg shadow-indigo-500/25">
              <svg className="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6A2.25 2.25 0 0 1 6 3.75h2.25A2.25 2.25 0 0 1 10.5 6v2.25a2.25 2.25 0 0 1-2.25 2.25H6a2.25 2.25 0 0 1-2.25-2.25V6ZM3.75 15.75A2.25 2.25 0 0 1 6 13.5h2.25a2.25 2.25 0 0 1 2.25 2.25V18a2.25 2.25 0 0 1-2.25 2.25H6A2.25 2.25 0 0 1 3.75 18v-2.25ZM13.5 6a2.25 2.25 0 0 1 2.25-2.25H18A2.25 2.25 0 0 1 20.25 6v2.25A2.25 2.25 0 0 1 18 10.5h-2.25a2.25 2.25 0 0 1-2.25-2.25V6ZM13.5 15.75a2.25 2.25 0 0 1 2.25-2.25H18a2.25 2.25 0 0 1 2.25 2.25V18A2.25 2.25 0 0 1 18 20.25h-2.25A2.25 2.25 0 0 1 13.5 18v-2.25Z" />
              </svg>
            </div>
            <span className="text-lg font-semibold text-gray-900 dark:text-white tracking-tight">Project1</span>
          </div>

          {/* Registration success banner */}
          {registered === "true" && (
            <div className="animate-in fade-in slide-in-from-top-4 rounded-xl border border-emerald-200 bg-emerald-50 p-4 dark:border-emerald-800 dark:bg-emerald-950">
              <div className="flex items-center gap-3">
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-900">
                  <svg className="h-4 w-4 text-emerald-600 dark:text-emerald-400" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                  </svg>
                </div>
                <div>
                  <p className="text-sm font-medium text-emerald-800 dark:text-emerald-200">
                    Registration successful!
                  </p>
                  <p className="text-xs text-emerald-600/80 dark:text-emerald-400/80 font-light">
                    Please sign in with your new account.
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Header */}
          <div className="space-y-1">
            <h2 className="text-2xl/[1.2] font-semibold tracking-tight text-gray-900 dark:text-white">
              Sign in
            </h2>
            <p className="text-sm text-gray-500 dark:text-gray-400 font-light">
              Enter your credentials to access your account.
            </p>
          </div>

          {/* Form */}
          <form className="space-y-5" onSubmit={handleSubmit} noValidate>
            <InputField
              label="Email address"
              id="email"
              type="email"
              autoComplete="email"
              placeholder="you@example.com"
              value={formData.email}
              onChange={handleChange}
              error={errors.email}
            />
            <InputField
              label="Password"
              id="password"
              type="password"
              autoComplete="current-password"
              placeholder="••••••••"
              value={formData.password}
              onChange={handleChange}
              error={errors.password}
            />

            {/* Links row */}
            <div className="flex items-center justify-between pt-1">
              <Link
                href="/forgot-password"
                className="text-sm font-medium text-indigo-600 transition-colors hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300"
              >
                Forgot password?
              </Link>
              <Link
                href="/register"
                className="text-sm font-medium text-indigo-600 transition-colors hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300"
              >
                Create account
              </Link>
            </div>

            {/* General error */}
            {errors.general && (
              <div className="animate-in fade-in rounded-lg border border-red-200 bg-red-50 p-3 dark:border-red-800 dark:bg-red-950">
                <p className="text-sm text-red-600 dark:text-red-400">{errors.general}</p>
              </div>
            )}

            {/* Submit button */}
            <button
              type="submit"
              disabled={isSubmitting}
              className={`group relative w-full inline-flex items-center justify-center rounded-lg px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 dark:focus:ring-offset-gray-950
                ${
                  isSubmitting
                    ? "cursor-not-allowed bg-indigo-400"
                    : "bg-linear-to-r from-indigo-600 to-indigo-700 hover:from-indigo-500 hover:to-indigo-600 hover:shadow-md hover:shadow-indigo-500/20 active:scale-[0.98]"
                }`}
            >
              {isSubmitting && (
                <svg className="mr-2 h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
              )}
              {isSubmitting ? "Signing in…" : "Sign in"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
