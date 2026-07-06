"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import Link from "next/link";
import type { Post } from "@/lib/types/post.types";

interface FeedViewProps {
  initialPosts: Post[];
  nextCursor: string | null;
  hasMore: boolean;
}

function truncate(s: string, n = 150): string {
  return s.length <= n ? s : s.slice(0, n) + "...";
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
      <div className="flex flex-col items-center justify-center rounded-2xl bg-white p-12 shadow-sm ring-1 ring-gray-200 dark:bg-gray-900 dark:ring-gray-800">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-gray-800 mb-4">
          <svg className="h-8 w-8 text-gray-400" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
          </svg>
        </div>
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white">No posts yet</h2>
        <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">Follow users or create a post to fill your feed.</p>
        <div className="mt-6 flex gap-3">
          <Link href="/posts/create" className="inline-flex items-center gap-2 rounded-lg bg-linear-to-r from-indigo-600 to-indigo-700 px-5 py-2.5 text-sm font-medium text-white shadow-sm transition-all hover:from-indigo-500 hover:to-indigo-600 hover:shadow-md active:scale-[0.98]">
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" /></svg>
            Create Post
          </Link>
          <Link href="/posts" className="inline-flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-5 py-2.5 text-sm font-medium text-gray-700 shadow-sm transition-all hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700">
            Browse Posts
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {posts.map((post) => (
        <article key={post.id} className="group rounded-2xl bg-white p-6 shadow-sm ring-1 ring-gray-200 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:ring-gray-300 dark:bg-gray-900 dark:ring-gray-800 dark:hover:ring-gray-700">
          <div className="flex items-center justify-between mb-3">
            {post.author ? (
              <Link href={`/${post.author}`} className="text-sm font-medium text-gray-700 transition-colors hover:text-indigo-600 dark:text-gray-300 dark:hover:text-indigo-400">{post.author}</Link>
            ) : <span className="text-sm text-gray-400 dark:text-gray-500">Unknown</span>}
            <span className="text-xs text-gray-400 dark:text-gray-500">{new Date(post.created_at).toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" })}</span>
          </div>
          <Link href={`/posts/${post.id}`} className="block">
            <h3 className="text-lg font-semibold tracking-tight text-gray-900 transition-colors group-hover:text-indigo-600 dark:text-white dark:group-hover:text-indigo-400">{post.title}</h3>
            <p className="mt-2 text-sm leading-relaxed text-gray-500 dark:text-gray-400">{truncate(post.content)}</p>
          </Link>
          <div className="mt-4 flex items-center justify-between">
            <div className="flex items-center gap-2 flex-wrap">
              {post.tags?.map((t) => <span key={t} className="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-600 dark:bg-gray-800 dark:text-gray-400">{t}</span>)}
            </div>
            <div className="flex items-center gap-3 text-xs">
              {post.like_count !== undefined && (
                <span className="flex items-center gap-1 text-gray-400 dark:text-gray-500">
                  <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" d="M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12z" /></svg>
                  {post.like_count}
                </span>
              )}
              <Link href={`/posts/${post.id}`} className="font-medium text-indigo-600 transition-transform group-hover:translate-x-0.5 dark:text-indigo-400">
                Read more →
              </Link>
            </div>
          </div>
        </article>
      ))}

      {more && <div ref={sentinel} data-feed-sentinel="true" className="h-4" />}

      {loading && (
        <div className="flex items-center justify-center gap-3 py-6 text-gray-500 dark:text-gray-400">
          <svg className="h-5 w-5 animate-spin" viewBox="0 0 24 24" fill="none"><circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" /><path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
          <span className="text-sm font-medium">Loading…</span>
        </div>
      )}

      {error && (
        <div className="flex items-center justify-between rounded-xl bg-red-50 px-5 py-4 ring-1 ring-red-200 dark:bg-red-950/30 dark:ring-red-800">
          <p className="text-sm text-red-700 dark:text-red-400">{error}</p>
          <button onClick={fetchNext} className="rounded-lg bg-red-100 px-3 py-1.5 text-sm font-medium text-red-700 hover:bg-red-200 dark:bg-red-900/50 dark:text-red-300 dark:hover:bg-red-900 transition-colors">Retry</button>
        </div>
      )}

      {!more && posts.length > 0 && !loading && (
        <div className="py-8 text-center"><p className="text-sm text-gray-400 dark:text-gray-500">You&apos;re all caught up ✨</p></div>
      )}
    </div>
  );
}
