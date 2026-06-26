import Link from "next/link";
import { redirect } from "next/navigation";
import { serverFetch } from "@/lib/server-fetch";
import { ApiError, handleApiResponse } from "@/lib/errors";
import { Post } from "@/lib/types/post.types";
import Navbar from "@/components/layout/navbar";

async function getPosts() {
  try {
    const res = await serverFetch("/api/posts");
    const posts = await handleApiResponse<Post[]>(res);
    return Array.isArray(posts) ? posts : [];
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      redirect("/login");
    }
    throw err;
  }
}

const truncateContent = (content: string, limit: number = 100) => {
  if (content.length <= limit) return content;
  return content.slice(0, limit) + "...";
};

export default async function PostsPage() {
  const posts = await getPosts();

  return (
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-gray-950">
      <Navbar pageTitle="Posts" />

      <main className="mx-auto w-full max-w-5xl flex-1 p-6 sm:p-8">
        {posts.length === 0 ? (
          <div className="flex flex-col items-center justify-center rounded-2xl bg-white p-12 shadow-sm ring-1 ring-gray-200 dark:bg-gray-900 dark:ring-gray-800">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-gray-800 mb-4">
              <svg className="h-8 w-8 text-gray-400" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white">No posts yet</h2>
            <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
              Start sharing your thoughts with the community.
            </p>
            <Link
              href="/posts/create"
              className="mt-6 inline-flex items-center gap-2 rounded-lg bg-linear-to-r from-indigo-600 to-indigo-700 px-5 py-2.5 text-sm font-medium text-white shadow-sm transition-all hover:from-indigo-500 hover:to-indigo-600 hover:shadow-md active:scale-[0.98]"
            >
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
              </svg>
              Create Your First Post
            </Link>
          </div>
        ) : (
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {posts.map((post) => (
              <Link
                key={post.id}
                href={`/posts/${post.id}`}
                className="group flex flex-col rounded-2xl bg-white p-6 shadow-sm ring-1 ring-gray-200 transition-all hover:shadow-lg hover:ring-gray-300 hover:-translate-y-0.5 dark:bg-gray-900 dark:ring-gray-800 dark:hover:ring-gray-700"
              >
                <h3 className="text-lg font-semibold tracking-tight text-gray-900 group-hover:text-indigo-600 dark:text-white dark:group-hover:text-indigo-400 transition-colors">
                  {post.title}
                </h3>
                <p className="mt-3 flex-1 text-sm leading-relaxed text-gray-500 dark:text-gray-400">
                  {truncateContent(post.content)}
                </p>
                <div className="mt-5 flex items-center justify-between text-xs text-gray-400 dark:text-gray-500">
                  <span>{new Date(post.created_at).toLocaleDateString(undefined, {
                    month: "short",
                    day: "numeric",
                    year: "numeric"
                  })}</span>
                  <span className="font-medium text-indigo-600 dark:text-indigo-400 group-hover:translate-x-0.5 transition-transform">
                    Read more →
                  </span>
                </div>
              </Link>
            ))}
          </div>
        )}
      </main>

      <footer className="py-6 text-center text-xs text-gray-400 dark:text-gray-600">
        &copy; {new Date().getFullYear()} Project One. All rights reserved.
      </footer>
    </div>
  );
}
