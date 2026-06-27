import { notFound } from "next/navigation";
import { serverFetch } from "@/lib/server-fetch";
import { ApiError, handleApiResponse } from "@/lib/errors";
import { Post } from "@/lib/types/post.types";
import PostInteractionSection from "@/components/posts/post-interaction-section";
import { cookies } from "next/headers";
import Navbar from "@/components/layout/navbar";

async function getPost(id: string) {
  try {
    const res = await serverFetch(`/api/posts/${id}`);
    const post = await handleApiResponse<Post>(res);
    if (!post) {
      notFound();
    }
    return post;
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 404) {
        notFound();
      }
    }
    throw err;
  }
}

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function PostDetailPage({ params }: PageProps) {
  const { id } = await params;
  const post = await getPost(id);
  const cookieStore = await cookies();
  const isAuthenticated = cookieStore.has("access_token");

  return (
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-gray-950">
      <Navbar pageTitle="Post Details" />

      <main className="mx-auto w-full max-w-3xl flex-1 p-6 sm:p-8">
        <article className="rounded-2xl bg-white p-6 shadow-sm ring-1 ring-gray-200 dark:bg-gray-900 dark:ring-gray-800 sm:p-10">
          <header className="mb-8 border-b border-gray-100 pb-8 dark:border-gray-800">
            <h1 className="text-3xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-4xl/[1.2]">
              {post.title}
            </h1>
            <div className="mt-4 flex flex-wrap items-center gap-3 text-sm text-gray-500 dark:text-gray-400">
              {post.author && (
                <>
                  <span className="inline-flex items-center gap-1.5 font-medium text-gray-700 dark:text-gray-300">
                    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                    </svg>
                    {post.author}
                  </span>
                  <span className="text-gray-300 dark:text-gray-700">·</span>
                </>
              )}
              <time dateTime={post.created_at}>
                {new Date(post.created_at).toLocaleDateString(undefined, {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                })}
              </time>
              {post.updated_at !== post.created_at && (
                <span className="italic text-gray-400">
                  (Updated: {new Date(post.updated_at).toLocaleDateString()})
                </span>
              )}
            </div>
            {post.tags && post.tags.length > 0 && (
              <div className="mt-5 flex flex-wrap gap-2">
                {post.tags.map((tag) => (
                  <span
                    key={tag}
                    className="inline-flex items-center rounded-full bg-indigo-50 px-3 py-1 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-700/10 dark:bg-indigo-950 dark:text-indigo-300 dark:ring-indigo-300/10"
                  >
                    #{tag}
                  </span>
                ))}
              </div>
            )}
          </header>

          <div className="prose prose-gray max-w-none dark:prose-invert">
            <p className="whitespace-pre-wrap text-base leading-relaxed text-gray-700 dark:text-gray-300">
              {post.content}
            </p>
          </div>

          <PostInteractionSection
            postId={post.id}
            postAuthor={post.author}
            initialComments={post.comments}
            isGuest={!isAuthenticated}
            initialLikeCount={post.like_count}
          />
        </article>
      </main>

      <footer className="py-6 text-center text-xs text-gray-400 dark:text-gray-600">
        &copy; {new Date().getFullYear()} Project One. All rights reserved.
      </footer>
    </div>
  );
}
