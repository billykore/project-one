import Link from "next/link";
import { serverFetch } from "@/lib/server-fetch";
import { handleApiResponse } from "@/lib/errors";
import type { SearchResponse } from "@/lib/types/search.types";

interface SearchPageProps {
  searchParams: Promise<{ q?: string }>;
}

export default async function SearchPage({ searchParams }: SearchPageProps) {
  const { q } = await searchParams;

  if (!q || q.trim().length === 0) {
    return (
      <main className="mx-auto max-w-2xl px-4 py-12">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">Search Users</h1>
        <p className="mt-4 text-gray-500 dark:text-gray-400">Enter a search term to find users.</p>
      </main>
    );
  }

  const res = await serverFetch(`/api/users/search?q=${encodeURIComponent(q)}&limit=20`);

  if (!res.ok) {
    // ponytail: generic error; per-error-type messaging if users complain
    return (
      <main className="mx-auto max-w-2xl px-4 py-12">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">
          Search results for &ldquo;{q}&rdquo;
        </h1>
        <p className="mt-4 text-gray-500 dark:text-gray-400">
          Something went wrong. Please try again.
        </p>
      </main>
    );
  }

  const { data: results } = await handleApiResponse<SearchResponse>(res);

  return (
    <main className="mx-auto max-w-2xl px-4 py-12">
      <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">
        Search results for &ldquo;{q}&rdquo;
      </h1>

      {results.length === 0 ? (
        <p className="mt-4 text-gray-500 dark:text-gray-400">
          No users found matching &ldquo;{q}&rdquo;.
        </p>
      ) : (
        <ul className="mt-6 divide-y divide-gray-200 dark:divide-gray-800">
          {results.map((user) => (
            <li key={user.username}>
              <Link
                href={`/${user.username}`}
                className="flex items-center gap-3 py-3 transition-colors hover:bg-gray-50 dark:hover:bg-gray-900 rounded-lg px-2 -mx-2"
              >
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-linear-to-br from-indigo-500 to-purple-600 text-sm font-medium text-white">
                  {user.name.charAt(0).toUpperCase()}
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-900 dark:text-white">
                    {user.name}
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    @{user.username}
                  </p>
                </div>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}
