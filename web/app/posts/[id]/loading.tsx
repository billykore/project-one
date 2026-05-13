export default function PostDetailLoading() {
  return (
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-black">
      <nav className="flex items-center justify-between bg-white px-8 py-4 shadow-sm dark:bg-zinc-900">
        <div className="flex items-center gap-4">
          <div className="h-5 w-24 animate-pulse rounded bg-gray-200 dark:bg-zinc-800" />
          <span className="text-xl font-bold text-gray-900 dark:text-zinc-50">Post Details</span>
        </div>
        <div className="flex items-center gap-4">
          <div className="h-9 w-28 animate-pulse rounded bg-gray-200 dark:bg-zinc-800" />
        </div>
      </nav>

      <main className="mx-auto w-full max-w-3xl flex-1 p-8">
        <div className="rounded-xl bg-white p-8 shadow-lg dark:bg-zinc-900 sm:p-12">
          <header className="mb-8 border-b border-gray-100 pb-8 dark:border-zinc-800">
            <div className="h-10 w-3/4 animate-pulse rounded bg-gray-200 dark:bg-zinc-800" />
            <div className="mt-4 flex gap-4">
              <div className="h-4 w-32 animate-pulse rounded bg-gray-100 dark:bg-zinc-800" />
            </div>
            <div className="mt-6 flex gap-2">
              <div className="h-6 w-16 animate-pulse rounded-full bg-indigo-50 dark:bg-indigo-950" />
              <div className="h-6 w-16 animate-pulse rounded-full bg-indigo-50 dark:bg-indigo-950" />
            </div>
          </header>

          <div className="space-y-4">
            <div className="h-4 w-full animate-pulse rounded bg-gray-100 dark:bg-zinc-800" />
            <div className="h-4 w-full animate-pulse rounded bg-gray-100 dark:bg-zinc-800" />
            <div className="h-4 w-5/6 animate-pulse rounded bg-gray-100 dark:bg-zinc-800" />
            <div className="h-4 w-full animate-pulse rounded bg-gray-100 dark:bg-zinc-800" />
            <div className="h-4 w-4/5 animate-pulse rounded bg-gray-100 dark:bg-zinc-800" />
          </div>
        </div>
      </main>

      <footer className="bg-white py-6 text-center text-sm text-gray-500 dark:bg-zinc-900 dark:text-zinc-400">
        &copy; {new Date().getFullYear()} Project One. All rights reserved.
      </footer>
    </div>
  );
}
