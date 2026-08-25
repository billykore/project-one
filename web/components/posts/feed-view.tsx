"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import Link from "next/link";
import type { Post } from "@/lib/types/post.types";

interface FeedViewProps {
  initialPosts: Post[];
  nextCursor: string | null;
  hasMore: boolean;
}

function truncate(s: string, n = 180): string {
  return s.length <= n ? s : s.slice(0, n).trimEnd() + "\u2026";
}

export function FeedView({ initialPosts, nextCursor, hasMore }: FeedViewProps) {
  const [posts, setPosts] = useState<Post[]>(initialPosts);
  const [cursor, setCursor] = useState<string | null>(nextCursor);
  const [more, setMore] = useState(hasMore);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fetching = useRef(false);
  const sentinel = useRef<HTMLDivElement | null>(null);

  const fetchNext = useCallback(async () => {
    if (fetching.current || !more || cursor === null) return;
    fetching.current = true;
    setLoading(true);
    setError(null);
    try {
      const searchParams = new URLSearchParams({ limit: "10" });
      searchParams.set("cursor", cursor);
      const res = await fetch(`/api/feeds?${searchParams.toString()}`);
      if (res.status === 401) { window.location.href = "/login"; return; }
      if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || `Failed (${res.status})`);
      const { data, next_cursor, has_more } = await res.json();
      setPosts((prev) => {
        const seenIds = new Set(prev.map((post) => post.id));
        return [...prev, ...(data ?? []).filter((post: Post) => !seenIds.has(post.id))];
      });
      setCursor(next_cursor || null);
      setMore(has_more ?? false);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load");
    } finally {
      setLoading(false);
      fetching.current = false;
    }
  }, [cursor, more]);

  useEffect(() => {
    const el = sentinel.current;
    if (!el || !more) return;
    const obs = new IntersectionObserver(([e]) => { if (e.isIntersecting) fetchNext(); }, { rootMargin: "200px" });
    obs.observe(el);
    return () => obs.disconnect();
  }, [fetchNext, more]);

  if (posts.length === 0 && !more && !loading) {
    return (
      <div className="panel px-6 py-16 text-center">
        <p className="meta">Nothing filed yet</p>
        <h2 className="mt-3 text-2xl font-semibold text-ink">Your feed is empty</h2>
        <p className="prose-lead measure mx-auto mt-2">
          Follow a few people, or write the first entry yourself.
        </p>
        <div className="mt-7 flex flex-wrap justify-center gap-2">
          <Link href="/posts/create" className="btn btn-primary btn-lg">
            Write a post
          </Link>
          <Link href="/posts" className="btn btn-quiet btn-lg">
            Browse everything
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {posts.map((post) => (
        <article key={post.id} className="entry grid gap-3 p-5 sm:grid-cols-[7rem_1fr] sm:gap-7 sm:p-6">
          {/* The margin: who wrote it, when, and how it landed. */}
          <div className="flex items-center gap-3 sm:flex-col sm:items-end sm:gap-1.5 sm:text-right">
            {post.author ? (
              <Link href={`/${post.author}`} className="meta font-medium text-ink hover:text-accent">
                {post.author}
              </Link>
            ) : (
              <span className="meta text-faint">Unknown</span>
            )}
            <time dateTime={post.created_at} className="meta">
              {new Date(post.created_at).toLocaleDateString(undefined, {
                day: "numeric",
                month: "short",
                year: "numeric",
              })}
            </time>
            {post.like_count !== undefined && post.like_count > 0 && (
              <span className="meta sm:mt-1" title={`${post.like_count} likes`}>
                <span aria-hidden="true" className="text-ember">&#9829;</span> {post.like_count}
              </span>
            )}
          </div>

          {/* The entry itself. */}
          <div>
            <h3 className="text-xl leading-snug font-semibold text-ink">
              <Link href={`/posts/${post.id}`} className="hover:text-accent">
                {post.title}
              </Link>
            </h3>
            <p className="prose-lead mt-2">{truncate(post.content)}</p>
            {post.tags && post.tags.length > 0 && (
              <div className="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1">
                {post.tags.map((t) => (
                  <span key={t} className="tag">{t}</span>
                ))}
              </div>
            )}
          </div>
        </article>
      ))}

      {more && <div ref={sentinel} data-feed-sentinel="true" className="h-4" />}

      {loading && (
        <p className="meta flex items-center justify-center gap-2 py-8">
          <svg className="h-3.5 w-3.5 animate-spin" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <circle className="opacity-30" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" />
            <path className="opacity-90" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          Loading more
        </p>
      )}

      {error && (
        <div className="notice notice-bad items-center justify-between">
          <p>{error}</p>
          <button onClick={fetchNext} className="btn btn-danger btn-sm">Try again</button>
        </div>
      )}

      {!more && posts.length > 0 && !loading && (
        <p className="rubric justify-center pt-6 before:h-px before:flex-1 before:bg-rule">
          End of feed
        </p>
      )}
    </div>
  );
}
