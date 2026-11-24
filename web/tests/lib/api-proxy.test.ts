import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { proxyToBackend } from "@/lib/api-proxy";

describe("proxyToBackend", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  function loginRequest() {
    return new Request("http://localhost:3000/api/login", {
      method: "POST",
      headers: { "Content-Type": "application/json", Cookie: "a=b" },
      body: JSON.stringify({ email: "user@example.com", password: "password123" }),
    });
  }

  it("forwards the status, problem body, and Retry-After of a rate-limited response", async () => {
    const problem = { status: 429, code: "RATE_LIMITED", detail: "Too many login attempts. Please try again later." };
    vi.mocked(global.fetch).mockResolvedValue(
      new Response(JSON.stringify(problem), {
        status: 429,
        headers: { "Content-Type": "application/problem+json", "Retry-After": "42" },
      }),
    );

    const res = await proxyToBackend(loginRequest(), "/auth/login");

    expect(res.status).toBe(429);
    expect(res.headers.get("Retry-After")).toBe("42");
    expect(await res.json()).toEqual(problem);
  });

  it("does not add Retry-After when the backend omits it", async () => {
    vi.mocked(global.fetch).mockResolvedValue(
      new Response(JSON.stringify({ message: "Login successful" }), { status: 200 }),
    );

    const res = await proxyToBackend(loginRequest(), "/auth/login");

    expect(res.status).toBe(200);
    expect(res.headers.get("Retry-After")).toBeNull();
  });
});
