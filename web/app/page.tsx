import Link from "next/link";
import { redirect } from "next/navigation";
import { serverFetch } from "@/lib/server-fetch";
import { ApiError, handleApiResponse } from "@/lib/errors";
import { FeedView } from "@/components/posts/FeedView";
import Navbar from "@/components/layout/navbar";
import type { Post } from "@/lib/types/post.types";

interface FeedApiResponse {
  data: Post[];
  has_more: boolean;
  next_cursor: string | null;
}

async function getFeed(): Promise<FeedApiResponse> {
  const res = await serverFetch("/api/feeds?limit=10");
  return handleApiResponse<FeedApiResponse>(res);
}

export default async function HomePage() {
  let feed: FeedApiResponse;
  try {
    feed = await getFeed();
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/login");
    return (
      <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-gray-950">
        <Navbar pageTitle="Home" />
        <main className="flex flex-1 items-center justify-center p-6">
          <div className="flex flex-col items-center gap-4 text-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30">
              <svg className="h-8 w-8 text-red-500" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white">Failed to load feed</h2>
            <p className="text-sm text-gray-500 dark:text-gray-400">Something went wrong. Please try again.</p>
            <Link href="/" className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 transition-colors">Retry</Link>
          </div>
        </main>
        <footer className="py-6 text-center text-xs text-gray-400 dark:text-gray-600">&copy; {new Date().getFullYear()} Project One. All rights reserved.</footer>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-gray-950">
      <Navbar pageTitle="Home" />
      <main className="mx-auto w-full max-w-2xl flex-1 p-6 sm:p-8">
        <FeedView initialPosts={feed.data} nextCursor={feed.next_cursor} hasMore={feed.has_more} />
      </main>
      <footer className="py-6 text-center text-xs text-gray-400 dark:text-gray-600">&copy; {new Date().getFullYear()} Project One. All rights reserved.</footer>
    </div>
  );
}
