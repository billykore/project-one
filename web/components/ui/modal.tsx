"use client";

import React, { useEffect, useRef } from "react";

interface ModalProps {
  open: boolean;
  onClose: () => void;
  /** Mono eyebrow naming what this dialog is about. */
  rubric: string;
  title: string;
  /** Tints the dialog's top edge. Destructive and failure states get ember. */
  tone?: "neutral" | "ember";
  children?: React.ReactNode;
  footer?: React.ReactNode;
  /** Focused when the dialog opens — always the safe choice, never the destructive one. */
  initialFocusRef?: React.RefObject<HTMLElement | null>;
  labelledBy?: string;
  describedBy?: string;
  className?: string;
}

/**
 * The one dialog shell in the app. Escape closes, the backdrop closes, focus
 * moves in on open and returns to wherever it came from on close.
 */
export function Modal({
  open,
  onClose,
  rubric,
  title,
  tone = "neutral",
  children,
  footer,
  initialFocusRef,
  labelledBy,
  describedBy,
  className = "",
}: ModalProps) {
  const returnFocusRef = useRef<HTMLElement | null>(null);
  const panelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [open, onClose]);

  useEffect(() => {
    if (open) {
      returnFocusRef.current = document.activeElement as HTMLElement;
      (initialFocusRef?.current ?? panelRef.current)?.focus();
    } else {
      returnFocusRef.current?.focus();
      returnFocusRef.current = null;
    }
  }, [open, initialFocusRef]);

  if (!open) return null;

  const titleId = labelledBy ?? "modal-title";

  return (
    <div
      className="enter fixed inset-0 z-50 flex items-center justify-center bg-ink/45 p-4 backdrop-blur-[2px]"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      aria-describedby={describedBy}
    >
      <div
        ref={panelRef}
        tabIndex={-1}
        onClick={(e) => e.stopPropagation()}
        className={`pop w-full max-w-md border-t-2 p-6 outline-none ${
          tone === "ember" ? "border-t-ember" : "border-t-accent"
        } ${className}`.trim()}
      >
        <p className="meta">{rubric}</p>
        <h3 id={titleId} className="mt-2 text-xl font-semibold text-ink">
          {title}
        </h3>
        {children}
        {footer && <div className="mt-7 flex flex-wrap justify-end gap-2">{footer}</div>}
      </div>
    </div>
  );
}

interface ConfirmDialogProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  rubric: string;
  title: string;
  description: string;
  /** Label of the button that performs the action — same verb the trigger used. */
  confirmLabel: string;
  /** Shown on the confirm button while the request is in flight. */
  pendingLabel: string;
  pending?: boolean;
  error?: string | null;
  tone?: "neutral" | "ember";
}

export function ConfirmDialog({
  open,
  onClose,
  onConfirm,
  rubric,
  title,
  description,
  confirmLabel,
  pendingLabel,
  pending = false,
  error = null,
  tone = "ember",
}: ConfirmDialogProps) {
  const cancelRef = useRef<HTMLButtonElement>(null);

  return (
    <Modal
      open={open}
      onClose={onClose}
      rubric={rubric}
      title={title}
      tone={tone}
      initialFocusRef={cancelRef}
      describedBy="modal-desc"
      footer={
        <>
          <button ref={cancelRef} onClick={onClose} disabled={pending} className="btn btn-quiet">
            Cancel
          </button>
          <button
            onClick={onConfirm}
            disabled={pending}
            className={`btn ${tone === "ember" ? "btn-danger" : "btn-primary"}`}
          >
            {pending && <Spinner />}
            {pending ? pendingLabel : confirmLabel}
          </button>
        </>
      }
    >
      <p id="modal-desc" className="prose-lead mt-3">
        {description}
      </p>
      {error && (
        <p className="mt-4 font-mono text-xs text-ember" role="alert">
          {error}
        </p>
      )}
    </Modal>
  );
}

export function Spinner({ className = "h-3.5 w-3.5" }: { className?: string }) {
  return (
    <svg className={`animate-spin ${className}`} viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <circle className="opacity-30" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" />
      <path className="opacity-90" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
    </svg>
  );
}
