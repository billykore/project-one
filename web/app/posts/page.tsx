"use client";

import Link from "next/link";
import Image from "next/image";
import { usePosts } from "./controller";

export default function PostsPage() {
  const { posts, isLoading, error, truncateContent } = usePosts();

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50">
        <div className="text-xl font-semibold text-gray-700">Loading posts...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50">
        <div className="text-xl font-semibold text-red-600">Error: {error}</div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-black">
      <nav className="flex items-center justify-between bg-white px-8 py-4 shadow-sm dark:bg-zinc-900">
        <div className="flex items-center gap-4">
          <Link href="/home">
            <Image
              className="dark:invert"
              src="/next.svg"
              alt="Next.js logo"
              width={100}
              height={20}
              priority
            />
          </Link>
          <span className="text-xl font-bold text-gray-900 dark:text-zinc-50">Posts</span>
        </div>
        <div className="flex items-center gap-4">
          <Link
            href="/posts/create"
            className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
          >
            Create New Post
          </Link>
          <Link
            href="/home"
            className="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-300 dark:hover:bg-zinc-700"
          >
            Dashboard
          </Link>
        </div>
      </nav>

      <main className="mx-auto w-full max-w-5xl flex-1 p-8">
        {posts.length === 0 ? (
          <div className="flex flex-col items-center justify-center rounded-xl bg-white p-12 shadow-lg dark:bg-zinc-900">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-zinc-50">No posts yet</h2>
            <p className="mt-4 text-gray-600 dark:text-zinc-400">
              You haven&apos;t created any posts. Start sharing your thoughts!
            </p>
            <Link
              href="/posts/create"
              className="mt-8 rounded-md bg-indigo-600 px-6 py-3 text-base font-medium text-white hover:bg-indigo-700 transition-colors"
            >
              Create Your First Post
            </Link>
          </div>
        ) : (
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {posts.length > 0 ? posts.map((post) => (
              <Link 
                key={post.id} 
                href={`/posts/${post.id}`}
                className="group flex flex-col rounded-xl bg-white p-6 shadow-md transition-all hover:shadow-xl dark:bg-zinc-900"
              >
                <h3 className="text-xl font-bold text-gray-900 group-hover:text-indigo-600 dark:text-zinc-50 dark:group-hover:text-indigo-400">
                  {post.title}
                </h3>
                <p className="mt-4 flex-1 text-gray-600 dark:text-zinc-400">
                  {truncateContent(post.content)}
                </p>
                <div className="mt-6 flex items-center justify-between text-sm text-gray-400">
                  <span>{new Date(post.created_at).toLocaleDateString()}</span>
                  <span className="font-medium text-indigo-600 dark:text-indigo-400 group-hover:underline">
                    Read more →
                  </span>
                </div>
              </Link>
            )) : <p>No posts available.</p>}
          </div>
        )}
      </main>

      <footer className="bg-white py-6 text-center text-sm text-gray-500 dark:bg-zinc-900 dark:text-zinc-400">
        &copy; {new Date().getFullYear()} Project One. All rights reserved.
      </footer>
    </div>
  );
}
