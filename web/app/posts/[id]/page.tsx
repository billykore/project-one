import Link from "next/link";
import Image from "next/image";
import { notFound } from "next/navigation";
import { apiServer } from "@/lib/api-server";
import { ApiError } from "@/lib/api";
import { Post } from "../model";
import PostInteractionSection from "@/components/posts/PostInteractionSection";
import DeletePostButton from "@/components/posts/DeletePostButton";
import { cookies } from "next/headers";

async function getPost(id: string) {
  try {
    const post = await apiServer.get<Post>(`/api/v1/posts/${id}`);
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
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-black">
      <nav className="flex items-center justify-between bg-white px-8 py-4 shadow-sm dark:bg-zinc-900">
        <div className="flex items-center gap-4">
          <Link href={isAuthenticated ? "/home" : "/login"}>
            <Image
              className="dark:invert"
              src="/next.svg"
              alt="Next.js logo"
              width={100}
              height={20}
              priority
            />
          </Link>
          <span className="text-xl font-bold text-gray-900 dark:text-zinc-50">Post Details</span>
        </div>
        <div className="flex items-center gap-4">
          {isAuthenticated ? (
            <>
              <DeletePostButton postId={post.id} postAuthor={post.author} />
              <Link
                href="/posts"
                className="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-300 dark:hover:bg-zinc-700"
              >
                Back to Posts
              </Link>
            </>
          ) : (
            <Link
              href="/login"
              className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 transition-colors"
            >
              Log In
            </Link>
          )}
        </div>
      </nav>

      <main className="mx-auto w-full max-w-3xl flex-1 p-8">
        <article className="rounded-xl bg-white p-8 shadow-lg dark:bg-zinc-900 sm:p-12">
          <header className="mb-8 border-b border-gray-100 pb-8 dark:border-zinc-800">
            <h1 className="text-4xl font-extrabold tracking-tight text-gray-900 dark:text-zinc-50">
              {post.title}
            </h1>
            <div className="mt-4 flex flex-wrap items-center gap-4 text-sm text-gray-500">
              {/* ponytail: simple author metadata display */}
              {post.author && (
                <>
                  <span className="font-semibold text-gray-700 dark:text-zinc-300">
                    Created by {post.author}
                  </span>
                  <span>•</span>
                </>
              )}
              <time dateTime={post.created_at}>
                Created on {new Date(post.created_at).toLocaleDateString(undefined, {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric'
                })}
              </time>
              {post.updated_at !== post.created_at && (
                <span className="italic">
                  (Updated: {new Date(post.updated_at).toLocaleDateString()})
                </span>
              )}
            </div>
            {post.tags && post.tags.length > 0 && (
              <div className="mt-6 flex flex-wrap gap-2">
                {post.tags.map((tag) => (
                  <span
                    key={tag}
                    className="rounded-full bg-indigo-50 px-3 py-1 text-xs font-semibold text-indigo-600 dark:bg-indigo-950 dark:text-indigo-400"
                  >
                    #{tag}
                  </span>
                ))}
              </div>
            )}
          </header>

          <div className="prose prose-indigo max-w-none dark:prose-invert">
            <p className="whitespace-pre-wrap text-lg leading-relaxed text-gray-700 dark:text-zinc-300">
              {post.content}
            </p>
          </div>

          <PostInteractionSection
            postId={post.id}
            initialComments={post.comments}
            isGuest={!isAuthenticated}
            initialLikeCount={post.like_count}
          />
        </article>
      </main>

      <footer className="bg-white py-6 text-center text-sm text-gray-500 dark:bg-zinc-900 dark:text-zinc-400">
        &copy; {new Date().getFullYear()} Project One. All rights reserved.
      </footer>
    </div>
  );
}
