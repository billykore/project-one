import React from "react";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { describe, expect, it, vi, beforeEach } from "vitest";
import ProfileDropdown from "@/components/layout/profile-dropdown";

// Mock localStorage
class LocalStorageMock {
  private store: Record<string, string> = {};

  clear() {
    this.store = {};
  }

  getItem(key: string) {
    return this.store[key] || null;
  }

  setItem(key: string, value: string) {
    this.store[key] = String(value);
  }

  removeItem(key: string) {
    delete this.store[key];
  }
}

const localStorageMock = new LocalStorageMock();
Object.defineProperty(global, "localStorage", {
  value: localStorageMock,
  writable: true,
});

const pushMock = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: pushMock,
  }),
}));

const fetchMock = vi.fn();
global.fetch = fetchMock;

vi.mock("@/lib/errors", () => ({
  ApiError: class ApiError extends Error {
    status: number;
    constructor(message: string, status: number) {
      super(message);
      this.status = status;
    }
  },
  handleApiResponse: (res: Response) => res.json(),
}));

describe("ProfileDropdown", () => {
  const mockUser = {
    name: "John Doe",
    username: "johndoe",
    email: "john@example.com",
  };

  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    localStorage.setItem("username", "johndoe");
  });

  it("renders the initials on the avatar button", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(<ProfileDropdown user={mockUser} />);
    });

    const button = container.querySelector("button[title='Account Menu']");
    expect(button).not.toBeNull();
    expect(button?.textContent).toBe("JD");
  });

  it("toggles the dropdown menu when clicked", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(<ProfileDropdown user={mockUser} />);
    });

    const button = container.querySelector("button[title='Account Menu']");
    
    // Dropdown should not be visible initially
    expect(container.querySelector("a[href='/home']")).toBeNull();

    // Click to open
    await act(async () => {
      button?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    });

    expect(container.querySelector("a[href='/home']")).not.toBeNull();
    expect(container.textContent).toContain("Signed in as");
    expect(container.textContent).toContain("John Doe");
    expect(container.textContent).toContain("@johndoe");

    // Click again to close
    await act(async () => {
      button?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    });

    expect(container.querySelector("a[href='/home']")).toBeNull();
  });

  it("opens logout confirmation modal when Log Out is clicked", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);

    await act(async () => {
      createRoot(container).render(<ProfileDropdown user={mockUser} />);
    });

    const button = container.querySelector("button[title='Account Menu']");
    
    // Open dropdown
    await act(async () => {
      button?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    });

    const logoutBtn = Array.from(container.querySelectorAll("button")).find(
      (btn) => btn.textContent?.includes("Log Out")
    );
    expect(logoutBtn).not.toBeUndefined();

    // Confirm dialog should not be open
    expect(container.querySelector("h3[id='modal-title']")).toBeNull();

    // Click Log Out
    await act(async () => {
      logoutBtn?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    });

    // Dropdown is closed, modal is open
    expect(container.querySelector("a[href='/home']")).toBeNull();
    expect(container.querySelector("h3[id='modal-title']")).not.toBeNull();
    expect(container.textContent).toContain("Confirm Logout");
  });

  it("triggers API logout and redirects to login when confirmed", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);
    fetchMock.mockResolvedValueOnce({ ok: true, json: async () => ({}) });

    await act(async () => {
      createRoot(container).render(<ProfileDropdown user={mockUser} />);
    });

    const button = container.querySelector("button[title='Account Menu']");
    
    // Open dropdown
    await act(async () => {
      button?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    });

    const logoutBtn = Array.from(container.querySelectorAll("button")).find(
      (btn) => btn.textContent?.includes("Log Out")
    );

    // Open confirmation modal
    await act(async () => {
      logoutBtn?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    });

    const confirmBtn = Array.from(container.querySelectorAll("button")).find(
      (btn) => btn.textContent === "Logout"
    );
    expect(confirmBtn).not.toBeUndefined();

    // Click confirm logout
    await act(async () => {
      confirmBtn?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    });

    expect(fetchMock).toHaveBeenCalledWith("/api/logout", { method: "POST" });
    expect(localStorage.getItem("username")).toBeNull();
    expect(pushMock).toHaveBeenCalledWith("/login");
  });
});
