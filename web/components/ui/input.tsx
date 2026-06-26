import React from "react";

interface InputFieldProps {
  label: string;
  id: string;
  type: string;
  value?: string;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
  error?: string | string[];
  placeholder?: string;
  autoComplete?: string;
}

export const InputField: React.FC<InputFieldProps> = ({
  label,
  id,
  type,
  value,
  onChange,
  error,
  placeholder,
  autoComplete,
}) => {
  const errorMessage = Array.isArray(error) ? error.join(", ") : error;
  
  return (
    <div className="space-y-1.5">
      <label
        htmlFor={id}
        className="block text-sm font-medium text-gray-700 dark:text-gray-300"
      >
        {label}
      </label>
      <input
        id={id}
        name={id}
        type={type}
        autoComplete={autoComplete}
        className={`block w-full rounded-lg border bg-white px-4 py-2.5 text-sm text-gray-900 placeholder-gray-400 shadow-sm transition-all duration-200 outline-none
          dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500
          ${
            error
              ? "border-red-400 focus:border-red-500 focus:ring-2 focus:ring-red-500/20"
              : "border-gray-200 hover:border-gray-300 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 dark:border-gray-700 dark:hover:border-gray-600 dark:focus:border-indigo-400"
          }`}
        placeholder={placeholder}
        value={value}
        onChange={onChange}
        aria-invalid={error ? "true" : "false"}
        aria-describedby={error ? `${id}-error` : undefined}
      />
      {errorMessage && (
        <p className="text-xs text-red-500 mt-1" id={`${id}-error`} aria-live="polite">
          {errorMessage}
        </p>
      )}
    </div>
  );
};
