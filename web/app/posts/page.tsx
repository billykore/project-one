import Link from "next/link";
import { redirect } from "next/navigation";
import { serverFetch } from "@/lib/server-fetch";
import { ApiError, handleApiResponse } from "@/lib/errors";
import { Post } from "@/lib/types/post.types";
import Navbar from "@/components/layout/navbar";
import SiteFooter from "@/components/layout/site-footer";

async function getPosts() {
  try {
    const res = await serverFetch("/api/posts");
    const page = await handleApiResponse<{ data: Post[] }>(res);
    return page.data ?? [];
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      redirect("/login");
    }
    throw err;
  }
}

const truncateContent = (content: string, limit: number = 140) => {
  if (content.length <= limit) return content;
  return content.slice(0, limit).trimEnd() + "\u2026";
};

export default async function PostsPage() {
  const posts = await getPosts();

  return (
    <div className="flex min-h-screen flex-col bg-paper">
      <Navbar pageTitle="Posts" />

      <main className="mx-auto w-full max-w-5xl flex-1 px-4 py-8 sm:px-6 sm:py-10">
        <div className="mb-6 flex flex-wrap items-baseline justify-between gap-3">
          <h1 className="text-2xl font-semibold text-ink">Everything published</h1>
          <p className="meta">{posts.length === 1 ? "1 post" : `${posts.length} posts`}</p>
        </div>

        {posts.length === 0 ? (
          <div className="panel px-6 py-16 text-center">
            <p className="meta">Nothing filed yet</p>
            <h2 className="mt-3 text-2xl font-semibold text-ink">No posts yet</h2>
            <p className="prose-lead measure mx-auto mt-2">
              Be the first to put something here.
            </p>
            <Link href="/posts/create" className="btn btn-primary btn-lg mt-7">
              Write the first post
            </Link>
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {posts.map((post) => (
              <Link key={post.id} href={`/posts/${post.id}`} className="entry flex flex-col p-5">
                <time dateTime={post.created_at} className="meta">
                  {new Date(post.created_at).toLocaleDateString(undefined, {
                    day: "numeric",
                    month: "short",
                    year: "numeric",
                  })}
                </time>
                <h2 className="mt-2.5 text-lg leading-snug font-semibold text-ink">{post.title}</h2>
                <p className="prose-lead mt-2 flex-1 text-sm">{truncateContent(post.content)}</p>
                {post.tags && post.tags.length > 0 && (
                  <div className="mt-4 flex flex-wrap gap-x-2.5 gap-y-1">
                    {post.tags.slice(0, 3).map((t) => (
                      <span key={t} className="tag">{t}</span>
                    ))}
                  </div>
                )}
              </Link>
            ))}
          </div>
        )}
      </main>

      <SiteFooter />
    </div>
  );
}
