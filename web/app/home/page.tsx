"use client";

import { useEffect } from "react";
import Image from "next/image";
import Link from "next/link";
import { useHome } from "./controller";

export default function HomePage() {
  const { 
    user, 
    isLoading, 
    error, 
    handleLogout, 
    isLogoutModalOpen, 
    isLoggingOut,
    openLogoutModal, 
    closeLogoutModal 
  } = useHome();

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isLogoutModalOpen) {
        closeLogoutModal();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isLogoutModalOpen, closeLogoutModal]);

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
      <nav className="flex items-center justify-between bg-white px-8 py-4 shadow-sm dark:bg-zinc-900">
        <Image
          className="dark:invert"
          src="/next.svg"
          alt="Next.js logo"
          width={100}
          height={20}
          priority
        />
        <div className="flex items-center gap-4">
          <Link
            href="/posts/create"
            className="rounded-md border border-indigo-600 px-4 py-2 text-sm font-medium text-indigo-600 hover:bg-indigo-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:text-indigo-400 dark:hover:bg-indigo-950"
          >
            Create New Post
          </Link>
          <button
            onClick={openLogoutModal}
            className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
          >
            Logout
          </button>
        </div>
      </nav>

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
                <Link
                  href="/posts"
                  className="rounded-md bg-white border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:bg-zinc-800 dark:border-zinc-700 dark:text-zinc-300 dark:hover:bg-zinc-700 transition-colors"
                >
                  View All Your Posts
                </Link>
                <Link
                  href="/posts/create"
                  className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 transition-colors"
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

      {/* Logout Confirmation Modal */}
      {isLogoutModalOpen && (
        <div 
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
          onClick={closeLogoutModal}
          role="dialog"
          aria-modal="true"
          aria-labelledby="modal-title"
          aria-describedby="modal-desc"
        >
          <div 
            className="w-full max-w-sm rounded-lg bg-white p-6 shadow-xl dark:bg-zinc-900"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 id="modal-title" className="text-xl font-bold text-gray-900 dark:text-zinc-50">Confirm Logout</h3>
            <p id="modal-desc" className="mt-4 text-gray-600 dark:text-zinc-400">
              Are you sure you want to log out? You will need to log back in to access your account.
            </p>
            <div className="mt-8 flex justify-end gap-3">
              <button
                onClick={closeLogoutModal}
                disabled={isLoggingOut}
                className="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none disabled:opacity-50 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-300 dark:hover:bg-zinc-700"
              >
                Cancel
              </button>
              <button
                onClick={handleLogout}
                disabled={isLoggingOut}
                className="inline-flex items-center rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 disabled:opacity-50"
              >
                {isLoggingOut ? (
                  <>
                    <svg className="-ml-1 mr-2 h-4 w-4 animate-spin text-white" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Logging out...
                  </>
                ) : (
                  "Logout"
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
