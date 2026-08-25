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
    <section className="panel w-full p-6">
      <h2 className="rubric">Password</h2>
      <p className="prose-lead mt-3 mb-5 text-sm">
        Change it whenever you like. You will stay signed in on this device.
      </p>

      <form onSubmit={submitPasswordChange} className="space-y-4" noValidate>
        <InputField
          label="Current password"
          id="oldPassword"
          type="password"
          value={pwdForm.oldPassword}
          onChange={handlePwdChange}
          error={errors.oldPassword}
        />
        <InputField
          label="New password"
          id="newPassword"
          type="password"
          value={pwdForm.newPassword}
          onChange={handlePwdChange}
          error={errors.newPassword}
        />
        <InputField
          label="Confirm new password"
          id="confirmPassword"
          type="password"
          value={pwdForm.confirmPassword}
          onChange={handlePwdChange}
          error={errors.confirmPassword}
        />

        {errors.general && (
          <p className="notice notice-bad" role="alert">
            {errors.general}
          </p>
        )}

        {pwdSuccess && (
          <p className="notice notice-good" role="status">
            {pwdSuccess}
          </p>
        )}

        <button type="submit" disabled={isSubmittingPwd} className="btn btn-primary btn-block">
          {isSubmittingPwd ? "Changing password" : "Change password"}
        </button>
      </form>
    </section>
  );
}
