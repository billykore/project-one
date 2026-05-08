import React from "react";

interface TextAreaFieldProps {
  label: string;
  id: string;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  error?: string;
  placeholder?: string;
  rows?: number;
}

export const TextAreaField: React.FC<TextAreaFieldProps> = ({
  label,
  id,
  value,
  onChange,
  error,
  placeholder,
  rows = 4,
}) => (
  <div>
    <label htmlFor={id} className="block text-sm font-medium text-gray-700 mb-1">
      {label}
    </label>
    <textarea
      id={id}
      name={id}
      rows={rows}
      required
      className={`appearance-none rounded-md relative block w-full px-3 py-2 border ${
        error ? "border-red-500" : "border-gray-300"
      } placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm`}
      placeholder={placeholder}
      value={value}
      onChange={onChange}
      aria-invalid={error ? "true" : "false"}
      aria-describedby={error ? `${id}-error` : undefined}
    />
    {error && (
      <p className="mt-2 text-sm text-red-600" id={`${id}-error`} aria-live="polite">
        {error}
      </p>
    )}
  </div>
);
