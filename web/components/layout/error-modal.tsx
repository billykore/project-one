"use client";

import React from "react";
import Link from "next/link";
import { useErrorModal } from "@/hooks/use-error-modal";
import { Modal } from "@/components/ui/modal";

export default function ErrorModal() {
  const { state, closeError } = useErrorModal();
  const { open, message, onRetry } = state;

  const handleRetry = () => {
    if (onRetry) {
      onRetry();
    } else {
      window.location.reload();
    }
    closeError();
  };

  return (
    <Modal
      open={open}
      onClose={closeError}
      rubric="Request failed"
      title="That didn't go through"
      tone="ember"
      labelledBy="error-modal-title"
      describedBy="error-modal-message"
      footer={
        <>
          <Link href="/" onClick={closeError} className="btn btn-quiet">
            Go home
          </Link>
          <button onClick={handleRetry} className="btn btn-primary">
            Try again
          </button>
        </>
      }
    >
      <p id="error-modal-message" className="prose-lead mt-3">
        {message}
      </p>
    </Modal>
  );
}
