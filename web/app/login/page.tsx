"use client";

import React, { useState, use } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ZodError } from "zod";
import { ApiError, ProblemDetail } from "@/lib/errors";
import { InputField } from "@/components/ui/input";
import { Spinner } from "@/components/ui/modal";
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
  searchParams: Promise<{ registered?: string; redirect?: string; reason?: string }>;
}) {
  const { registered, redirect: redirectUrl, reason } = use(searchParams);
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
        const errBody = await res.json().catch(() => ({}));
        const problem: ProblemDetail = {
          type: "about:blank",
          title: "Login Failed",
          status: res.status,
          detail: errBody.detail || errBody.error || "Invalid email or password. Please try again.",
          instance: "/login",
        };
        throw new ApiError(problem);
      }

      router.push(redirectUrl || "/");
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
    <div className="flex min-h-screen flex-col bg-paper px-4 py-10 sm:px-6">
      <div className="mx-auto flex w-full max-w-sm flex-1 flex-col justify-center">
        <header className="border-b border-rule pb-6">
          <p className="font-body text-2xl leading-none font-semibold tracking-tight text-ink">
            Project One
          </p>
          <p className="meta mt-3">A place to write things down</p>
        </header>

        {reason === "session_expired" && (
          <p className="notice notice-notice enter mt-6" role="status">
            Your session ran out. Sign in again to pick up where you left off.
          </p>
        )}

        {registered === "true" && (
          <p className="notice notice-good enter mt-6" role="status">
            Your account is ready. Sign in to get started.
          </p>
        )}

        <h1 className="mt-8 text-3xl font-semibold text-ink">Welcome back</h1>

        <form className="mt-7 space-y-5" onSubmit={handleSubmit} noValidate>
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
            placeholder="Your password"
            value={formData.password}
            onChange={handleChange}
            error={errors.password}
            showPasswordToggle
          />

          {errors.general && (
            <p className="notice notice-bad enter" role="alert">
              {errors.general}
            </p>
          )}

          <button type="submit" disabled={isSubmitting} className="btn btn-primary btn-lg btn-block">
            {isSubmitting && <Spinner />}
            {isSubmitting ? "Signing in" : "Sign in"}
          </button>
        </form>

        <div className="mt-6 flex items-baseline justify-between border-t border-rule pt-5">
          <Link href="/register" className="link text-sm font-medium">
            Create an account
          </Link>
          <Link href="/forgot-password" className="meta hover:text-accent">
            Forgot password
          </Link>
        </div>
      </div>

      <footer className="mx-auto w-full max-w-sm pt-10">
        <p className="meta text-faint">&copy; {new Date().getFullYear()} Project One</p>
      </footer>
    </div>
  );
}
