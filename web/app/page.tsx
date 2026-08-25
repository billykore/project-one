import Link from "next/link";
import { redirect } from "next/navigation";
import { serverFetch } from "@/lib/server-fetch";
import { ApiError, handleApiResponse } from "@/lib/errors";
import { FeedView } from "@/components/posts/feed-view";
import Navbar from "@/components/layout/navbar";
import SiteFooter from "@/components/layout/site-footer";
import type { Post } from "@/lib/types/post.types";

interface FeedApiResponse {
  data: Post[];
  has_more: boolean;
  next_cursor: string;
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
      <div className="flex min-h-screen flex-col bg-paper">
        <Navbar pageTitle="Home" />
        <main className="mx-auto w-full max-w-3xl flex-1 px-4 py-8 sm:px-6 sm:py-10">
          <div className="panel border-t-2 border-t-ember p-8">
            <p className="meta">Feed unavailable</p>
            <h1 className="mt-2 text-2xl font-semibold text-ink">We couldn&apos;t load your feed</h1>
            <p className="prose-lead mt-3">The server didn&apos;t answer. Try again in a moment.</p>
            <Link href="/" className="btn btn-primary mt-6">
              Try again
            </Link>
          </div>
        </main>
        <SiteFooter />
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col bg-paper">
      <Navbar pageTitle="Home" />
      <main className="mx-auto w-full max-w-3xl flex-1 px-4 py-8 sm:px-6 sm:py-10">
        <div className="mb-6 flex flex-wrap items-baseline justify-between gap-3">
          <h1 className="text-2xl font-semibold text-ink">Your feed</h1>
          <p className="meta">Newest first</p>
        </div>
        <FeedView
          initialPosts={feed.data}
          nextCursor={feed.next_cursor || null}
          hasMore={feed.has_more}
        />
      </main>
      <SiteFooter />
    </div>
  );
}
