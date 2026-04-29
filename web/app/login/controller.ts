import { useState } from "react";
import { loginSchema, LoginFormData, LoginErrors } from "./model";
import { ZodError } from "zod";

export const useLogin = () => {
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
        err.errors.forEach((e) => {
          const path = e.path[0] as keyof LoginFormData;
          if (!fieldErrors[path]) {
            fieldErrors[path] = e.message;
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

    try {
      // Mock API call
      console.log("Logging in with:", { email: formData.email.trim(), password: formData.password });
      
      await new Promise((resolve) => setTimeout(resolve, 1500));
      
      alert("Login attempt submitted successfully (Mock)");
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
