import React from "react";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { FeedView } from "@/components/posts/FeedView";

vi.mock("next/link", () => ({
  default: ({ children, ...props }: { children: React.ReactNode; [key: string]: unknown }) =>
    React.createElement("a", props, children),
}));

const observeMock = vi.fn();
const disconnectMock = vi.fn();
let intersectionCallback: IntersectionObserverCallback | undefined;

class MockIntersectionObserver implements IntersectionObserver {
  readonly root = null;
  readonly rootMargin = "200px";
  readonly thresholds: ReadonlyArray<number> = [];

  constructor(callback: IntersectionObserverCallback) {
    intersectionCallback = callback;
  }

  disconnect = disconnectMock;
  observe = observeMock;
  takeRecords = () => [];
  unobserve = vi.fn();
}

Object.defineProperty(global, "IntersectionObserver", {
  writable: true,
  value: MockIntersectionObserver,
});

function makePost(id: number, overrides: Partial<React.ComponentProps<typeof FeedView>["initialPosts"][number]> = {}) {
  return {
    id,
    title: `Post ${id}`,
    content: `Content for post ${id}`,
    author: `author${id}`,
    created_at: "2026-07-05T12:00:00.000Z",
    updated_at: "2026-07-05T12:00:00.000Z",
    tags: ["feed"],
    like_count: id,
    ...overrides,
  };
}

async function flushPromises() {
  await new Promise((resolve) => setTimeout(resolve, 0));
}

async function renderFeed(props: React.ComponentProps<typeof FeedView>) {
  const container = document.createElement("div");
  document.body.appendChild(container);
  const root = createRoot(container);

  await act(async () => {
    root.render(<FeedView {...props} />);
  });

  return { container, root };
}

describe("FeedView", () => {
  beforeEach(() => {
    observeMock.mockReset();
    disconnectMock.mockReset();
    intersectionCallback = undefined;
    vi.unstubAllGlobals();
  });

  it("renders the empty state actions when no posts are available", async () => {
    const { container } = await renderFeed({
      initialPosts: [],
      nextCursor: null,
      hasMore: false,
    });

    expect(container.textContent).toContain("No posts yet");
    expect(container.querySelector('a[href="/posts/create"]')).not.toBeNull();
    expect(container.querySelector('a[href="/posts"]')).not.toBeNull();
    expect(container.querySelector("[data-feed-sentinel='true']")).toBeNull();
  });

  it("observes the sentinel and loads the next page with an encoded cursor", async () => {
    const fetchMock = vi.fn().mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          data: [makePost(2, { title: "Loaded post" })],
          next_cursor: null,
          has_more: false,
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    const { container, root } = await renderFeed({
      initialPosts: [makePost(1)],
      nextCursor: "abc+=&",
      hasMore: true,
    });

    const sentinel = container.querySelector("[data-feed-sentinel='true']");
    expect(sentinel).not.toBeNull();
    expect(observeMock).toHaveBeenCalledWith(sentinel);

    await act(async () => {
      intersectionCallback?.(
        [{ isIntersecting: true, target: sentinel } as IntersectionObserverEntry],
        {} as IntersectionObserver,
      );
      await flushPromises();
    });

    expect(fetchMock).toHaveBeenCalledWith("/api/feeds?limit=10&cursor=abc%2B%3D%26");
    expect(container.textContent).toContain("Loaded post");
    expect(container.textContent).toContain("You're all caught up");

    await act(async () => {
      root.unmount();
    });
    expect(disconnectMock).toHaveBeenCalledTimes(1);
  });

  it("shows an inline error and retries loading more posts", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ error: "Feed unavailable" }), {
          status: 500,
          headers: { "Content-Type": "application/json" },
        }),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: [makePost(2, { title: "Recovered post" })],
            next_cursor: null,
            has_more: false,
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
      );
    vi.stubGlobal("fetch", fetchMock);

    const { container } = await renderFeed({
      initialPosts: [makePost(1)],
      nextCursor: "retry-cursor",
      hasMore: true,
    });

    const sentinel = container.querySelector("[data-feed-sentinel='true']");

    await act(async () => {
      intersectionCallback?.(
        [{ isIntersecting: true, target: sentinel } as IntersectionObserverEntry],
        {} as IntersectionObserver,
      );
      await flushPromises();
    });

    expect(container.textContent).toContain('Feed unavailable');

    const retryButton = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent === "Retry",
    );
    expect(retryButton).not.toBeUndefined();

    await act(async () => {
      retryButton?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
      await flushPromises();
    });

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(container.textContent).toContain("Recovered post");
    expect(container.textContent).not.toContain("Feed unavailable");
  });
});
