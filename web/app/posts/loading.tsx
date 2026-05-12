export default function PostsLoading() {
  return (
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-black">
      <nav className="flex items-center justify-between bg-white px-8 py-4 shadow-sm dark:bg-zinc-900">
        <div className="flex items-center gap-4">
          <div className="h-5 w-24 animate-pulse rounded bg-gray-200 dark:bg-zinc-800" />
          <span className="text-xl font-bold text-gray-900 dark:text-zinc-50">Posts</span>
        </div>
        <div className="flex items-center gap-4">
          <div className="h-9 w-32 animate-pulse rounded bg-indigo-200 dark:bg-indigo-900" />
          <div className="h-9 w-24 animate-pulse rounded bg-gray-200 dark:bg-zinc-800" />
        </div>
      </nav>

      <main className="mx-auto w-full max-w-5xl flex-1 p-8">
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {[...Array(6)].map((_, i) => (
            <div 
              key={i}
              className="flex flex-col rounded-xl bg-white p-6 shadow-md dark:bg-zinc-900"
            >
              <div className="h-7 w-3/4 animate-pulse rounded bg-gray-200 dark:bg-zinc-800" />
              <div className="mt-4 space-y-2">
                <div className="h-4 w-full animate-pulse rounded bg-gray-100 dark:bg-zinc-800" />
                <div className="h-4 w-5/6 animate-pulse rounded bg-gray-100 dark:bg-zinc-800" />
              </div>
              <div className="mt-6 flex items-center justify-between">
                <div className="h-4 w-20 animate-pulse rounded bg-gray-100 dark:bg-zinc-800" />
                <div className="h-4 w-16 animate-pulse rounded bg-indigo-100 dark:bg-indigo-950" />
              </div>
            </div>
          ))}
        </div>
      </main>

      <footer className="bg-white py-6 text-center text-sm text-gray-500 dark:bg-zinc-900 dark:text-zinc-400">
        &copy; {new Date().getFullYear()} Project One. All rights reserved.
      </footer>
    </div>
  );
}
