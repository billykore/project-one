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
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-black">
      <Navbar pageTitle="Posts" />

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
            {posts.map((post) => (
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
            ))}
          </div>
        )}
      </main>

      <footer className="bg-white py-6 text-center text-sm text-gray-500 dark:bg-zinc-900 dark:text-zinc-400">
        &copy; {new Date().getFullYear()} Project One. All rights reserved.
      </footer>
    </div>
  );
}
