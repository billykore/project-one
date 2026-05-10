import { useState } from "react";
import { loginSchema, LoginFormData, LoginErrors, LoginResponse } from "./model";
import { ZodError } from "zod";
import { api, ApiError } from "@/lib/api";
import { useRouter } from "next/navigation";

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

  const handleSubmit = async (e: React.SubmitEvent) => {
    e.preventDefault();
    
    if (!validate()) return;

    setIsSubmitting(true);
    setErrors({});

    try {
      await api.post<LoginResponse>("/users/login", {
        email: formData.email.trim(),
        password: formData.password,
      });

      // Redirect to home page (cookies are set by backend)
      router.push("/home");
    } catch (err) {
      if (err instanceof ApiError && (err.status === 401 || err.status === 400)) {
        setErrors({ general: err.message });
        return;
      }

      const errorMessage = err instanceof Error ? err.message : "An error occurred. Please try again later.";
      router.push(`/error?message=${encodeURIComponent(errorMessage)}`);
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
