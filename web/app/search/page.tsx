import Link from "next/link";
import { serverFetch } from "@/lib/server-fetch";
import { handleApiResponse } from "@/lib/errors";
import Navbar from "@/components/layout/navbar";
import SiteFooter from "@/components/layout/site-footer";
import { Avatar } from "@/components/ui/avatar";
import type { SearchResponse } from "@/lib/types/search.types";

interface SearchPageProps {
  searchParams: Promise<{ q?: string }>;
}

function SearchLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col bg-paper">
      <Navbar pageTitle="Search" />
      <main className="mx-auto w-full max-w-2xl flex-1 px-4 py-8 sm:px-6 sm:py-10">{children}</main>
      <SiteFooter />
    </div>
  );
}

export default async function SearchPage({ searchParams }: SearchPageProps) {
  const { q } = await searchParams;

  if (!q || q.trim().length === 0) {
    return (
      <SearchLayout>
        <p className="meta">Search</p>
        <h1 className="mt-2 text-2xl font-semibold text-ink">Find someone</h1>
        <p className="prose-lead mt-2">
          Type a name or username into the search box above.
        </p>
      </SearchLayout>
    );
  }

  const res = await serverFetch(`/api/users/search?q=${encodeURIComponent(q)}&limit=20`);

  if (!res.ok) {
    // ponytail: generic error; per-error-type messaging if users complain
    return (
      <SearchLayout>
        <p className="meta">Results for &ldquo;{q}&rdquo;</p>
        <h1 className="mt-2 text-2xl font-semibold text-ink">Search is unavailable</h1>
        <p className="prose-lead mt-2">The server didn&apos;t answer. Try again in a moment.</p>
      </SearchLayout>
    );
  }

  const { data: results } = await handleApiResponse<SearchResponse>(res);

  return (
    <SearchLayout>
      <p className="meta">
        {results.length === 1 ? "1 result" : `${results.length} results`} for &ldquo;{q}&rdquo;
      </p>
      <h1 className="mt-2 text-2xl font-semibold text-ink">People</h1>

      {results.length === 0 ? (
        <p className="prose-lead mt-3">
          Nobody here matches &ldquo;{q}&rdquo;. Try a different spelling.
        </p>
      ) : (
        <ul className="mt-6 divide-y divide-rule border-t border-rule">
          {results.map((user) => (
            <li key={user.username}>
              <Link
                href={`/${user.username}`}
                className="-mx-3 flex items-center gap-3 rounded-sm px-3 py-3.5 transition-colors hover:bg-sunken"
              >
                <Avatar name={user.name} />
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium text-ink">{user.name}</p>
                  <p className="meta truncate normal-case">@{user.username}</p>
                </div>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </SearchLayout>
  );
}
