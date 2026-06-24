import React from "react";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { ErrorModalProvider, useErrorModal } from "@/hooks/use-error-modal";
import ErrorModal from "@/components/layout/error-modal";
import Link from "next/dist/client/link";

// Mock next/link
vi.mock("next/link", () => ({
  default: ({
    children,
    onClick,
    className,
  }: {
    children: React.ReactNode;
    href: string;
    onClick?: () => void;
    className?: string;
  }) => (
    <Link href="/" onClick={onClick} className={className}>
      {children}
    </Link>
  ),
}));

// Mock window.location.reload
const reloadMock = vi.fn();
vi.stubGlobal("location", { reload: reloadMock });

function TestHarness({ children }: { children: React.ReactNode }) {
  return <ErrorModalProvider>{children}</ErrorModalProvider>;
}

function TriggerButton() {
  const { showError } = useErrorModal();
  return (
    <button onClick={() => showError("Test error")}>
      Trigger Error
    </button>
  );
}

describe("ErrorModal", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("does not render when no error is set", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
        </TestHarness>
      );
    });

    expect(container.querySelector('[role="dialog"]')).toBeNull();
  });

  it("renders modal with message when showError is called", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
          <TriggerButton />
        </TestHarness>
      );
    });

    await act(async () => {
      (container.querySelector("button") as HTMLButtonElement).click();
    });

    const dialog = container.querySelector('[role="dialog"]');
    expect(dialog).not.toBeNull();
    expect(dialog?.getAttribute("aria-modal")).toBe("true");
    expect(container.textContent).toContain("Test error");
    expect(container.textContent).toContain("Something went wrong");
    expect(container.textContent).toContain("Try again");
    expect(container.textContent).toContain("Go back home");
  });

  it("closes modal when backdrop is clicked", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
          <TriggerButton />
        </TestHarness>
      );
    });

    await act(async () => {
      (container.querySelector("button") as HTMLButtonElement).click();
    });

    const backdrop = container.querySelector('[role="dialog"]') as HTMLElement;
    await act(async () => {
      backdrop.click();
    });

    expect(container.querySelector('[role="dialog"]')).toBeNull();
  });

  it("does not close when dialog inner is clicked", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
          <TriggerButton />
        </TestHarness>
      );
    });

    await act(async () => {
      (container.querySelector("button") as HTMLButtonElement).click();
    });

    const inner = container.querySelector(".max-w-md") as HTMLElement;
    await act(async () => {
      inner.click();
    });

    expect(container.querySelector('[role="dialog"]')).not.toBeNull();
  });

  it("closes modal when Escape key is pressed", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
          <TriggerButton />
        </TestHarness>
      );
    });

    await act(async () => {
      (container.querySelector("button") as HTMLButtonElement).click();
    });

    await act(async () => {
      window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    });

    expect(container.querySelector('[role="dialog"]')).toBeNull();
  });

  it("calls onRetry callback when Try again is clicked", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    const retryFn = vi.fn();

    function TriggerWithRetry() {
      const { showError } = useErrorModal();
      return (
        <button onClick={() => showError("msg", retryFn)}>
          Trigger
        </button>
      );
    }

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
          <TriggerWithRetry />
        </TestHarness>
      );
    });

    await act(async () => {
      (container.querySelector("button") as HTMLButtonElement).click();
    });

    const tryAgainBtn = Array.from(container.querySelectorAll("button")).find(
      (b) => b.textContent === "Try again"
    ) as HTMLButtonElement;

    await act(async () => {
      tryAgainBtn.click();
    });

    expect(retryFn).toHaveBeenCalledTimes(1);
    expect(container.querySelector('[role="dialog"]')).toBeNull();
  });

  it("reloads page when Try again clicked and no onRetry provided", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
          <TriggerButton />
        </TestHarness>
      );
    });

    await act(async () => {
      (container.querySelector("button") as HTMLButtonElement).click();
    });

    const tryAgainBtn = Array.from(container.querySelectorAll("button")).find(
      (b) => b.textContent === "Try again"
    ) as HTMLButtonElement;

    await act(async () => {
      tryAgainBtn.click();
    });

    expect(reloadMock).toHaveBeenCalledTimes(1);
  });

  it("navigates home and closes when Go back home is clicked", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(
        <TestHarness>
          <ErrorModal />
          <TriggerButton />
        </TestHarness>
      );
    });

    await act(async () => {
      (container.querySelector("button") as HTMLButtonElement).click();
    });

    const homeLink = container.querySelector('a[href="/"]') as HTMLAnchorElement;

    await act(async () => {
      homeLink.click();
    });

    expect(container.querySelector('[role="dialog"]')).toBeNull();
  });
});
