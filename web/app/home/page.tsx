"use client";

import Link from "next/link";
import { useHome } from "./controller";
import Navbar from "../../components/layout/Navbar";

export default function HomePage() {
  const { 
    user, 
    isLoading, 
    error 
  } = useHome();

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50">
        <div className="text-xl font-semibold text-gray-700">Loading...</div>
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
      <Navbar pageTitle="Dashboard" />

      <main className="flex flex-1 flex-col items-center justify-center p-8">
        <div className="w-full max-w-3xl rounded-xl bg-white p-12 shadow-lg dark:bg-zinc-900">
          <div className="flex flex-col items-center gap-8 text-center sm:items-start sm:text-left">
            <h1 className="text-4xl font-bold tracking-tight text-gray-900 dark:text-zinc-50">
              Welcome, {user?.name}
            </h1>
            <p className="text-lg leading-8 text-gray-600 dark:text-zinc-400">
              You have successfully logged in to the User Service. 
              Your email is <span className="font-semibold text-gray-900 dark:text-zinc-50">{user?.email}</span>.
            </p>
            
            <div className="mt-8 border-t border-gray-200 pt-8 dark:border-zinc-800">
              <h2 className="text-xl font-semibold text-gray-900 dark:text-zinc-50">
                What&apos;s next?
              </h2>
              <p className="mt-4 text-gray-600 dark:text-zinc-400">
                This is your personal dashboard. From here you can manage your account and access all the features of our platform.
              </p>
              <div className="mt-6 flex flex-wrap gap-4 justify-center sm:justify-start">
                {user?.username && (
                  <Link
                    href={`/${user.username}`}
                    className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 transition-colors"
                  >
                    View Your Profile
                  </Link>
                )}
                <Link
                  href="/posts"
                  className="rounded-md bg-white border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:bg-zinc-800 dark:border-zinc-700 dark:text-zinc-300 dark:hover:bg-zinc-700 transition-colors"
                >
                  View All Your Posts
                </Link>
                <Link
                  href="/posts/create"
                  className={user?.username ? "rounded-md bg-white border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:bg-zinc-800 dark:border-zinc-700 dark:text-zinc-300 dark:hover:bg-zinc-700 transition-colors" : "rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 transition-colors"}
                >
                  Create a New Post
                </Link>
              </div>
            </div>
          </div>
        </div>
      </main>

      <footer className="bg-white py-6 text-center text-sm text-gray-500 dark:bg-zinc-900 dark:text-zinc-400">
        &copy; {new Date().getFullYear()} Project One. All rights reserved.
      </footer>


    </div>
  );
}
