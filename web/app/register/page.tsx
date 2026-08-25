"use client";

import React, { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ZodError } from "zod";
import { ApiError, type ProblemDetail } from "@/lib/errors";
import { InputField } from "@/components/ui/input";
import { Spinner } from "@/components/ui/modal";
import { useErrorModal } from "@/hooks/use-error-modal";
import { registerRequestBodySchema, RegisterRequestBody } from "@/app/api/register/schema";

interface RegisterErrors {
  first_name?: string;
  last_name?: string;
  username?: string;
  email?: string;
  password?: string;
  confirmPassword?: string;
  general?: string;
}

export default function RegisterPage() {
  const router = useRouter();
  const { showError } = useErrorModal();
  const [formData, setFormData] = useState<RegisterRequestBody>({
    first_name: "",
    last_name: "",
    username: "",
    email: "",
    password: "",
    confirmPassword: "",
  });
  const [errors, setErrors] = useState<RegisterErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    if (errors[name as keyof RegisterErrors]) {
      setErrors((prev) => ({ ...prev, [name]: undefined }));
    }
  };

  const validate = (): boolean => {
    try {
      registerRequestBodySchema.parse(formData);
      setErrors({});
      return true;
    } catch (err) {
      if (err instanceof ZodError) {
        const fieldErrors: RegisterErrors = {};
        err.issues.forEach((e) => {
          if (e.path.length > 0) {
            const path = e.path[0] as keyof RegisterErrors;
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
      const res = await fetch("/api/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          first_name: formData.first_name.trim(),
          last_name: formData.last_name.trim(),
          username: formData.username.trim(),
          email: formData.email.trim(),
          password: formData.password,
        }),
      });

      if (!res.ok) {
        const problem = (await res.json().catch(() => ({}))) as Partial<ProblemDetail>;
        throw new ApiError({
          type: problem.type ?? "about:blank",
          title: problem.title ?? "Registration failed",
          status: problem.status ?? res.status,
          detail: problem.detail ?? "Registration failed",
          instance: problem.instance ?? "",
          code: problem.code,
          errors: problem.errors,
        });
      }

      router.push("/login?registered=true");
    } catch (err) {
      if (err instanceof ApiError && err.status === 400) {
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

        <h1 className="mt-8 text-3xl font-semibold text-ink">Create an account</h1>
        <p className="prose-lead mt-2">
          Your name and username are shown on everything you post. Your email stays private.
        </p>

        <form className="mt-7 space-y-5" onSubmit={handleSubmit} noValidate>
          <div className="grid gap-5 sm:grid-cols-2">
            <InputField
              label="First name"
              id="first_name"
              type="text"
              autoComplete="given-name"
              placeholder="Ada"
              value={formData.first_name}
              onChange={handleChange}
              error={errors.first_name}
            />
            <InputField
              label="Last name"
              id="last_name"
              type="text"
              autoComplete="family-name"
              placeholder="Lovelace"
              value={formData.last_name}
              onChange={handleChange}
              error={errors.last_name}
            />
          </div>
          <InputField
            label="Username"
            id="username"
            type="text"
            autoComplete="username"
            placeholder="ada"
            value={formData.username}
            onChange={handleChange}
            error={errors.username}
          />
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
            autoComplete="new-password"
            placeholder="At least 8 characters"
            value={formData.password}
            onChange={handleChange}
            error={errors.password}
            showPasswordToggle
          />
          <InputField
            label="Confirm password"
            id="confirmPassword"
            type="password"
            autoComplete="new-password"
            placeholder="Type it again"
            value={formData.confirmPassword}
            onChange={handleChange}
            error={errors.confirmPassword}
          />

          {errors.general && (
            <p className="notice notice-bad enter" role="alert">
              {errors.general}
            </p>
          )}

          <button type="submit" disabled={isSubmitting} className="btn btn-primary btn-lg btn-block">
            {isSubmitting && <Spinner />}
            {isSubmitting ? "Creating account" : "Create account"}
          </button>
        </form>

        <p className="prose-lead mt-6 border-t border-rule pt-5 text-sm">
          Already have an account?{" "}
          <Link href="/login" className="link font-medium">
            Sign in
          </Link>
        </p>
      </div>

      <footer className="mx-auto w-full max-w-sm pt-10">
        <p className="meta text-faint">&copy; {new Date().getFullYear()} Project One</p>
      </footer>
    </div>
  );
}
