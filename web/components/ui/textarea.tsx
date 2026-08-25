import React from "react";

export type TextareaProps = React.TextareaHTMLAttributes<HTMLTextAreaElement>;

export const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, ...props }, ref) => {
    return (
      <textarea
        className={`field field-prose min-h-20 ${className || ""}`}
        ref={ref}
        {...props}
      />
    );
  }
);
Textarea.displayName = "Textarea";


interface TextAreaFieldProps {
  label: string;
  id: string;
  value?: string;
  onChange?: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  error?: string | string[];
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
}) => {
  const errorMessage = Array.isArray(error) ? error.join(", ") : error;

  return (
    <div className="space-y-1.5">
      <label htmlFor={id} className="meta block">
        {label}
      </label>
      <textarea
        id={id}
        name={id}
        rows={rows}
        className="field field-prose"
        placeholder={placeholder}
        defaultValue={value}
        onChange={onChange}
        aria-invalid={error ? "true" : "false"}
        aria-describedby={error ? `${id}-error` : undefined}
      />
      {errorMessage && (
        <p className="mt-1 font-mono text-xs text-ember" id={`${id}-error`} aria-live="polite">
          {errorMessage}
        </p>
      )}
    </div>
  );
};
