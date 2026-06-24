import React from "react";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// Mock next/navigation
const mockPush = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush }),
}));

// Mock next/link
vi.mock("next/link", () => ({
  default: ({ children, ...props }: { children: React.ReactNode; [key: string]: unknown }) =>
    React.createElement("a", props, children),
}));

// Mock the error modal hook
const mockShowError = vi.fn();
vi.mock("@/hooks/use-error-modal", () => ({
  useErrorModal: () => ({ showError: mockShowError }),
}));

let LoginPage: React.ComponentType;
let container: HTMLDivElement;

describe("LoginPage", () => {
  beforeEach(async () => {
    mockPush.mockReset();
    mockShowError.mockReset();
    vi.stubGlobal("fetch", vi.fn());
    container = document.createElement("div");
    document.body.appendChild(container);

    const mod = await import("@/app/(auth)/login/page");
    LoginPage = mod.default;
  });

  afterEach(() => {
    document.body.removeChild(container);
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  async function fillAndSubmit(email: string, password: string) {
    const root = createRoot(container);
    await act(async () => {
      root.render(React.createElement(LoginPage));
    });

    const emailInput = container.querySelector("#email") as HTMLInputElement;
    const passwordInput = container.querySelector("#password") as HTMLInputElement;
    const submitButton = container.querySelector('button[type="submit"]') as HTMLButtonElement;

    await act(async () => {
      // Use React's synthetic event via the native input setter + dispatch
      const emailSetter = Object.getOwnPropertyDescriptor(
        HTMLInputElement.prototype, "value"
      )?.set;
      emailSetter?.call(emailInput, email);
      emailInput.dispatchEvent(new Event("input", { bubbles: true }));

      const pwdSetter = Object.getOwnPropertyDescriptor(
        HTMLInputElement.prototype, "value"
      )?.set;
      pwdSetter?.call(passwordInput, password);
      passwordInput.dispatchEvent(new Event("input", { bubbles: true }));
    });

    await act(async () => {
      submitButton.click();
    });
  }

  it("calls /api/login on form submit", async () => {
    const mockFetch = vi.mocked(global.fetch);
    mockFetch.mockResolvedValue(
      new Response(JSON.stringify({ message: "ok", username: "testuser" }), { status: 200 }),
    );

    await fillAndSubmit("test@example.com", "password123");

    expect(mockFetch).toHaveBeenCalledWith(
      "/api/login",
      expect.objectContaining({
        method: "POST",
        headers: { "Content-Type": "application/json" },
      }),
    );

    const body = JSON.parse(mockFetch.mock.calls[0][1]!.body as string);
    expect(body).toEqual({ email: "test@example.com", password: "password123" });
  });

  it("redirects to / on successful login", async () => {
    const mockFetch = vi.mocked(global.fetch);
    mockFetch.mockResolvedValue(
      new Response(JSON.stringify({ message: "ok", username: "testuser" }), { status: 200 }),
    );

    await fillAndSubmit("test@example.com", "password123");

    // Wait for the async submit handler to resolve
    await act(async () => {
      await new Promise((r) => setTimeout(r, 0));
    });

    expect(mockPush).toHaveBeenCalledWith("/");
  });
});
