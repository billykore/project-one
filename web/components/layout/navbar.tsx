"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { handleApiResponse } from "@/lib/errors";
import NotificationDropdown from "@/components/notification/notification-dropdown";
import ProfileDropdown from "@/components/layout/profile-dropdown";
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
    <nav className="sticky top-0 z-40 flex items-center justify-between bg-white/80 backdrop-blur-md px-6 py-3 shadow-sm ring-1 ring-gray-200 dark:bg-gray-950/80 dark:ring-gray-800">
      {/* Left side: Logo & Dynamic page title */}
      <div className="flex items-center">
        <Link href={user ? "/" : "/login"} className="flex items-center gap-2.5">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-linear-to-br from-indigo-600 to-purple-700 shadow-sm">
            <svg className="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6A2.25 2.25 0 0 1 6 3.75h2.25A2.25 2.25 0 0 1 10.5 6v2.25a2.25 2.25 0 0 1-2.25 2.25H6a2.25 2.25 0 0 1-2.25-2.25V6ZM3.75 15.75A2.25 2.25 0 0 1 6 13.5h2.25a2.25 2.25 0 0 1 2.25 2.25V18a2.25 2.25 0 0 1-2.25 2.25H6A2.25 2.25 0 0 1 3.75 18v-2.25ZM13.5 6a2.25 2.25 0 0 1 2.25-2.25H18A2.25 2.25 0 0 1 20.25 6v2.25A2.25 2.25 0 0 1 18 10.5h-2.25a2.25 2.25 0 0 1-2.25-2.25V6ZM13.5 15.75a2.25 2.25 0 0 1 2.25-2.25H18a2.25 2.25 0 0 1 2.25 2.25V18A2.25 2.25 0 0 1 18 20.25h-2.25A2.25 2.25 0 0 1 13.5 18v-2.25Z" />
            </svg>
          </div>
          <span className="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">Project1</span>
        </Link>
        {pageTitle && (
          <>
            <span className="mx-3 h-5 w-px bg-gray-200 dark:bg-gray-700" />
            <span className="text-sm font-medium text-gray-500 dark:text-gray-400">
              {pageTitle}
            </span>
          </>
        )}
      </div>

      {/* Right side: Nav links and dropdowns */}
      <div className="flex items-center gap-3">
        {isLoading ? (
          <NavbarActionsSkeleton />
        ) : user ? (
          // Authenticated state
          <>
            {rightActions}
            <Link
              href="/posts/create"
              className="inline-flex items-center gap-1.5 rounded-lg bg-linear-to-r from-indigo-600 to-indigo-700 px-3.5 py-2 text-xs font-medium text-white shadow-sm transition-all hover:from-indigo-500 hover:to-indigo-600 hover:shadow-md active:scale-[0.98]"
            >
              <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" strokeWidth="2.5" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
              </svg>
              Create Post
            </Link>
            <NotificationDropdown />
            <ProfileDropdown user={user} />
          </>
        ) : (
          // Unauthenticated/Guest state
          <Link
            href="/login"
            className="inline-flex items-center rounded-lg bg-linear-to-r from-indigo-600 to-indigo-700 px-4 py-2 text-xs font-medium text-white shadow-sm transition-all hover:from-indigo-500 hover:to-indigo-600 hover:shadow-md active:scale-[0.98]"
          >
            Log In
          </Link>
        )}
      </div>
    </nav>
  );
}
