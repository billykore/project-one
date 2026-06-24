"use client";

import React, { createContext, useContext, useState, useCallback } from "react";

interface ErrorModalState {
  open: boolean;
  message: string;
  onRetry?: () => void;
}

interface ErrorModalContextValue {
  showError: (message: string, onRetry?: () => void) => void;
  closeError: () => void;
  state: ErrorModalState;
}

const ErrorModalContext = createContext<ErrorModalContextValue | null>(null);

export function ErrorModalProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<ErrorModalState>({
    open: false,
    message: "",
    onRetry: undefined,
  });

  const showError = useCallback((message: string, onRetry?: () => void) => {
    if (!message) return;
    setState({ open: true, message, onRetry });
  }, []);

  const closeError = useCallback(() => {
    setState({ open: false, message: "", onRetry: undefined });
  }, []);

  return (
    <ErrorModalContext.Provider value={{ showError, closeError, state }}>
      {children}
    </ErrorModalContext.Provider>
  );
}

export function useErrorModal(): ErrorModalContextValue {
  const ctx = useContext(ErrorModalContext);
  if (!ctx) {
    throw new Error("useErrorModal must be used within an ErrorModalProvider");
  }
  return ctx;
}
