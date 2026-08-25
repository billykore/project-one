"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { handleApiResponse } from "@/lib/errors";
import NotificationDropdown from "@/components/notification/notification-dropdown";
import ProfileDropdown from "@/components/layout/profile-dropdown";
import SearchBar from "@/components/layout/search-bar";
import { NavbarActionsSkeleton } from "@/components/ui/page-skeleton";

interface User {
  name: string;
  username: string;
  email: string;
}

interface NavbarProps {
  pageTitle?: string;
  rightActions?: React.ReactNode;
}

export default function Navbar({ pageTitle, rightActions }: NavbarProps) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadUser = async () => {
      try {
        // ponytail: read cookie instead of localStorage
        const username = document.cookie.split("; ").find(r => r.startsWith("username="))?.split("=")[1];
        if (!username) {
          setIsLoading(false);
          return;
        }

        const userData = await handleApiResponse<User>(await fetch(`/api/users/${username}`));
        setUser(userData);
      } catch (err) {
        console.error("Navbar failed to load user info:", err);
        // Do not force redirect; user might be a guest visiting a public page
      } finally {
        setIsLoading(false);
      }
    };

    loadUser();
  }, []);

  return (
    <nav className="sticky top-0 z-40 border-b border-rule bg-paper/92 backdrop-blur-sm">
      <div className="mx-auto flex w-full max-w-6xl items-center gap-4 px-4 py-3 sm:px-6">
        {/* Masthead: the journal's name, set in the reading serif. The flanks
            both take flex-1 from a zero basis, so they claim equal width and
            the search between them lands dead centre of the bar. */}
        <div className="flex min-w-0 flex-1 items-baseline gap-2.5">
          <Link
            href={user ? "/" : "/login"}
            className="font-body text-lg leading-none font-semibold tracking-tight whitespace-nowrap text-ink"
          >
            Project One
          </Link>
          {pageTitle && (
            <span className="hidden min-w-0 items-baseline gap-2.5 lg:flex">
              <span aria-hidden="true" className="meta text-faint">/</span>
              <span className="meta truncate">{pageTitle}</span>
            </span>
          )}
        </div>

        {/* Widens with the viewport, but never crowds the flanks. */}
        <div className="hidden w-full shrink md:block max-w-sm">
          <SearchBar />
        </div>

        <div className="flex flex-1 items-center justify-end gap-1.5">
          {isLoading ? (
            <NavbarActionsSkeleton />
          ) : user ? (
            <>
              {rightActions}
              <Link href="/posts/create" className="btn btn-sm btn-primary">
                Write
              </Link>
              <NotificationDropdown />
              <ProfileDropdown user={user} />
            </>
          ) : (
            <Link href="/login" className="btn btn-primary">
              Log in
            </Link>
          )}
        </div>
      </div>

      {/* Search collapses below the masthead on narrow screens. */}
      <div className="border-t border-rule px-4 py-2 md:hidden">
        <SearchBar />
      </div>
    </nav>
  );
}
