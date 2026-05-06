"use client";

import Image from "next/image";
import { useHome } from "./controller";

export default function HomePage() {
  const { 
    user, 
    isLoading, 
    error, 
    handleLogout, 
    isLogoutModalOpen, 
    openLogoutModal, 
    closeLogoutModal 
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
      <nav className="flex items-center justify-between bg-white px-8 py-4 shadow-sm dark:bg-zinc-900">
        <Image
          className="dark:invert"
          src="/next.svg"
          alt="Next.js logo"
          width={100}
          height={20}
          priority
        />
        <button
          onClick={openLogoutModal}
          className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
        >
          Logout
        </button>
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
        >
          <div 
            className="w-full max-w-sm rounded-lg bg-white p-6 shadow-xl dark:bg-zinc-900"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 className="text-xl font-bold text-gray-900 dark:text-zinc-50">Confirm Logout</h3>
            <p className="mt-4 text-gray-600 dark:text-zinc-400">
              Are you sure you want to log out? You will need to log back in to access your account.
            </p>
            <div className="mt-8 flex justify-end gap-3">
              <button
                onClick={closeLogoutModal}
                className="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-300 dark:hover:bg-zinc-700"
              >
                Cancel
              </button>
              <button
                onClick={handleLogout}
                className="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
