"use client";

import Link from "next/link";
import Navbar from "@/components/layout/navbar";
import { useUser } from "@/hooks/use-user";

export default function HomePage() {
  const { user, isLoading } = useUser();

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50 dark:bg-gray-950">
        <div className="flex items-center gap-3 text-gray-500 dark:text-gray-400">
          <svg className="h-5 w-5 animate-spin" viewBox="0 0 24 24" fill="none">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          <span className="text-sm font-medium">Loading…</span>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-gray-950">
      <Navbar pageTitle="Dashboard" />

      <main className="flex flex-1 flex-col items-center justify-center p-6 sm:p-8">
        <div className="w-full max-w-2xl space-y-8">
          {/* Welcome card */}
          <div className="rounded-2xl bg-white p-8 shadow-sm ring-1 ring-gray-200 dark:bg-gray-900 dark:ring-gray-800 sm:p-10">
            <div className="flex flex-col items-center gap-6 text-center">
              {/* Avatar */}
              <div className="flex h-16 w-16 items-center justify-center rounded-full bg-linear-to-br from-indigo-600 to-purple-700 text-xl font-bold uppercase tracking-wider text-white shadow-lg shadow-indigo-500/25">
                {user?.name ? user.name.split(" ").map((n: string) => n[0]).slice(0, 2).join("") : "?"}
              </div>

              <div className="space-y-2">
                <h1 className="text-2xl font-semibold tracking-tight text-gray-900 dark:text-white">
                  Welcome back, {user?.name}
                </h1>
                <p className="text-sm text-gray-500 dark:text-gray-400 max-w-sm mx-auto leading-relaxed">
                  You&apos;re signed in as{" "}
                  <span className="font-medium text-gray-700 dark:text-gray-300">{user?.email}</span>.
                  This is your personal dashboard.
                </p>
              </div>

              {/* Action buttons */}
              <div className="flex flex-wrap items-center justify-center gap-3 pt-2">
                {user?.username && (
                  <Link
                    href={`/${user.username}`}
                    className="inline-flex items-center gap-2 rounded-lg bg-linear-to-r from-indigo-600 to-indigo-700 px-4 py-2.5 text-sm font-medium text-white shadow-sm transition-all hover:from-indigo-500 hover:to-indigo-600 hover:shadow-md hover:shadow-indigo-500/20 active:scale-[0.98]"
                  >
                    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                    </svg>
                    View Profile
                  </Link>
                )}
                <Link
                  href="/posts"
                  className="inline-flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 shadow-sm transition-all hover:bg-gray-50 hover:border-gray-300 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
                >
                  <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 4a2 2 0 00-2-2m2 2v10a2 2 0 01-2 2M9 9h6m-6 4h6m-6 4h5" />
                  </svg>
                  All Posts
                </Link>
                <Link
                  href="/posts/create"
                  className="inline-flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 shadow-sm transition-all hover:bg-gray-50 hover:border-gray-300 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
                >
                  <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
                  </svg>
                  Create Post
                </Link>
              </div>
            </div>
          </div>
        </div>
      </main>

      <footer className="py-6 text-center text-xs text-gray-400 dark:text-gray-600">
        &copy; {new Date().getFullYear()} Project One. All rights reserved.
      </footer>
    </div>
  );
}
