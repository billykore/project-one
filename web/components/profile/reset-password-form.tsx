"use client";

import React from "react";
import { InputField } from "@/components/ui/input";
import { ChangePasswordFormData } from "@/lib/types/profile.types";

interface ResetPasswordFormProps {
  pwdForm: ChangePasswordFormData;
  errors: Record<string, string>;
  isSubmittingPwd: boolean;
  pwdSuccess: string | null;
  handlePwdChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  submitPasswordChange: (e: React.FormEvent) => void;
}

export default function ResetPasswordForm({
  pwdForm,
  errors,
  isSubmittingPwd,
  pwdSuccess,
  handlePwdChange,
  submitPasswordChange,
}: ResetPasswordFormProps) {
  return (
    <div className="w-full rounded-xl border border-zinc-200 bg-white p-6 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-zinc-50">Security</h3>
      <p className="text-sm text-zinc-500 dark:text-zinc-400 mb-6">
        Update your password to keep your account secure.
      </p>

      <form onSubmit={submitPasswordChange} className="space-y-4" noValidate>
        <InputField
          label="Current Password"
          id="oldPassword"
          type="password"
          value={pwdForm.oldPassword}
          onChange={handlePwdChange}
          error={errors.oldPassword}
        />
        <InputField
          label="New Password"
          id="newPassword"
          type="password"
          value={pwdForm.newPassword}
          onChange={handlePwdChange}
          error={errors.newPassword}
        />
        <InputField
          label="Confirm New Password"
          id="confirmPassword"
          type="password"
          value={pwdForm.confirmPassword}
          onChange={handlePwdChange}
          error={errors.confirmPassword}
        />

        {errors.general && (
          <div className="rounded-md bg-red-50 p-3 text-sm text-red-700 dark:bg-red-900/10 dark:text-red-400">
            {errors.general}
          </div>
        )}

        {pwdSuccess && (
          <div className="rounded-md bg-green-50 p-3 text-sm text-green-700 dark:bg-green-900/10 dark:text-green-400">
            {pwdSuccess}
          </div>
        )}

        <button
          type="submit"
          disabled={isSubmittingPwd}
          className="w-full rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700 transition disabled:opacity-50 cursor-pointer"
        >
          {isSubmittingPwd ? "Updating..." : "Change Password"}
        </button>
      </form>
    </div>
  );
}
