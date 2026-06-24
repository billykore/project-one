import React from "react";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { describe, expect, it } from "vitest";
import { ErrorModalProvider, useErrorModal } from "@/hooks/use-error-modal";

function TestConsumer({
  onRender,
}: {
  onRender: (ctx: ReturnType<typeof useErrorModal>) => void;
}) {
  const ctx = useErrorModal();
  onRender(ctx);
  return null;
}

describe("useErrorModal", () => {
  it("sets open=true and message when showError is called", async () => {
    let captured: ReturnType<typeof useErrorModal> | null = null;
    const capture = (ctx: ReturnType<typeof useErrorModal>) => { captured = ctx; };
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <ErrorModalProvider>
          <TestConsumer onRender={capture} />
        </ErrorModalProvider>
      );
    });

    await act(async () => {
      captured!.showError("Test error message");
    });

    expect(captured!.state.open).toBe(true);
    expect(captured!.state.message).toBe("Test error message");
  });

  it("does not open when showError is called with empty message", async () => {
    let captured: ReturnType<typeof useErrorModal> | null = null;
    const capture = (ctx: ReturnType<typeof useErrorModal>) => { captured = ctx; };
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <ErrorModalProvider>
          <TestConsumer onRender={capture} />
        </ErrorModalProvider>
      );
    });

    await act(async () => {
      captured!.showError("");
    });

    expect(captured!.state.open).toBe(false);
  });

  it("stores onRetry callback when provided", async () => {
    let captured: ReturnType<typeof useErrorModal> | null = null;
    const capture = (ctx: ReturnType<typeof useErrorModal>) => { captured = ctx; };
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <ErrorModalProvider>
          <TestConsumer onRender={capture} />
        </ErrorModalProvider>
      );
    });

    const retryFn = () => {};
    await act(async () => {
      captured!.showError("msg", retryFn);
    });

    expect(captured!.state.onRetry).toBe(retryFn);
  });

  it("closeError resets state to closed", async () => {
    let captured: ReturnType<typeof useErrorModal> | null = null;
    const capture = (ctx: ReturnType<typeof useErrorModal>) => { captured = ctx; };
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <ErrorModalProvider>
          <TestConsumer onRender={capture} />
        </ErrorModalProvider>
      );
    });

    await act(async () => {
      captured!.showError("msg", () => {});
    });
    expect(captured!.state.open).toBe(true);

    await act(async () => {
      captured!.closeError();
    });

    expect(captured!.state.open).toBe(false);
    expect(captured!.state.message).toBe("");
    expect(captured!.state.onRetry).toBeUndefined();
  });

  it("replaces existing error when showError is called while open", async () => {
    let captured: ReturnType<typeof useErrorModal> | null = null;
    const capture = (ctx: ReturnType<typeof useErrorModal>) => { captured = ctx; };
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <ErrorModalProvider>
          <TestConsumer onRender={capture} />
        </ErrorModalProvider>
      );
    });

    const firstRetry = () => {};
    const secondRetry = () => {};

    await act(async () => {
      captured!.showError("first", firstRetry);
    });
    expect(captured!.state.message).toBe("first");

    await act(async () => {
      captured!.showError("second", secondRetry);
    });
    expect(captured!.state.message).toBe("second");
    expect(captured!.state.onRetry).toBe(secondRetry);
    expect(captured!.state.open).toBe(true);
  });
});
