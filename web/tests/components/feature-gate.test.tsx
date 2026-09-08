import React, { type ReactNode } from "react";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { describe, expect, it, vi } from "vitest";
import { FeatureGate } from "@/components/feature-flag/feature-gate";

function render(node: ReactNode) {
  const container = document.createElement("div");
  document.body.appendChild(container);
  act(() => {
    createRoot(container).render(node);
  });
  return container;
}

describe("FeatureGate", () => {
  it("renders children when the backend enables the flag", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({
      environment: "local",
      decisions: [{ key: "editor", enabled: true, source: "enabled_all" }],
    }), { status: 200, headers: { "Content-Type": "application/json" } })));
    const container = render(<FeatureGate flag="editor"><span>editor</span></FeatureGate>);
    await act(async () => {});
    expect(container.textContent).toBe("editor");
    vi.unstubAllGlobals();
  });

  it("renders the fallback when the backend disables the flag", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({
      environment: "local",
      decisions: [{ key: "editor", enabled: false, source: "disabled_all" }],
    }), { status: 200, headers: { "Content-Type": "application/json" } })));
    const container = render(<FeatureGate flag="editor" fallback={<span>unavailable</span>}><span>editor</span></FeatureGate>);
    await act(async () => {});
    expect(container.textContent).toBe("unavailable");
    vi.unstubAllGlobals();
  });
});
