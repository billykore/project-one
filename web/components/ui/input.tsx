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
    <div>
      <label htmlFor={id} className="block text-sm font-medium text-gray-700 mb-1">
        {label}
      </label>
      <input
        id={id}
        name={id}
        type={type}
        autoComplete={autoComplete}
        className={`appearance-none rounded-md relative block w-full px-3 py-2 border ${
          error ? "border-red-500" : "border-gray-300"
        } placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm`}
        placeholder={placeholder}
        value={value}
        onChange={onChange}
        aria-invalid={error ? "true" : "false"}
        aria-describedby={error ? `${id}-error` : undefined}
      />
      {errorMessage && (
        <p className="mt-2 text-sm text-red-600" id={`${id}-error`} aria-live="polite">
          {errorMessage}
        </p>
      )}
    </div>
  );
};
