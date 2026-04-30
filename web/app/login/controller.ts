import { useState } from "react";
import { loginSchema, LoginFormData, LoginErrors, LoginResponse } from "./model";
import { ZodError } from "zod";
import { api } from "@/lib/api";
import { useRouter } from "next/navigation";

interface RawLoginResponse {
  access_token: string;
  refresh_token: string;
}

const setSecureCookie = (name: string, value: string, days: number) => {
  const expires = new Date();
  expires.setTime(expires.getTime() + days * 24 * 60 * 60 * 1000);
  // Using SameSite=Strict and Secure for better client-side security
  document.cookie = `${name}=${value};expires=${expires.toUTCString()};path=/;SameSite=Strict;Secure`;
};

export const useLogin = () => {
  const router = useRouter();
  const [formData, setFormData] = useState<LoginFormData>({
    email: "",
    password: "",
  });
  const [errors, setErrors] = useState<LoginErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    
    // Clear error for the field when user starts typing
    if (errors[name as keyof LoginFormData]) {
      setErrors((prev) => ({ ...prev, [name]: undefined }));
    }
  };

  const validate = (): boolean => {
    try {
      loginSchema.parse(formData);
      setErrors({});
      return true;
    } catch (err) {
      if (err instanceof ZodError) {
        const fieldErrors: LoginErrors = {};
        err.issues.forEach((e) => {
          if (e.path.length > 0) {
            const path = e.path[0] as keyof LoginFormData;
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
      const response = await api.post<RawLoginResponse>("/user/login", {
        email: formData.email.trim(),
        password: formData.password,
      });

      const loginData: LoginResponse = {
        accessToken: response.access_token,
        refreshToken: response.refresh_token,
      };

      // Store tokens securely on the client side
      setSecureCookie("access_token", loginData.accessToken, 1); // 1 day
      setSecureCookie("refresh_token", loginData.refreshToken, 7); // 7 days

      // Redirect to home page
      router.push("/");
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "An error occurred. Please try again later.";
      setErrors({ general: errorMessage });
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
    formData,
    errors,
    isSubmitting,
    handleChange,
    handleSubmit,
  };
};
