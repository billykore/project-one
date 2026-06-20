"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import Image from "next/image";
import { api } from "@/lib/api";
import NotificationDropdown from "../notification/NotificationDropdown";
import ProfileDropdown from "./ProfileDropdown";

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
        const username = localStorage.getItem("username");
        if (!username) {
          setIsLoading(false);
          return;
        }

        const userData = await api.get<User>(`/api/v1/users/${username}`);
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
    <nav className="flex items-center justify-between bg-white px-8 py-4 shadow-sm dark:bg-zinc-900 border-b border-gray-100 dark:border-zinc-800">
      {/* Left side: Logo & Dynamic page title */}
      <div className="flex items-center">
        <Link href={user ? "/home" : "/login"}>
          <Image
            className="dark:invert"
            src="/next.svg"
            alt="Next.js logo"
            width={90}
            height={18}
            priority
          />
        </Link>
        {pageTitle && (
          <>
            <span className="h-5 w-px bg-zinc-200 dark:bg-zinc-700 mx-4"></span>
            <span className="text-md font-bold text-gray-900 dark:text-zinc-50">
              {pageTitle}
            </span>
          </>
        )}
      </div>

      {/* Right side: Nav links and dropdowns */}
      <div className="flex items-center gap-4">
        {isLoading ? (
          // Skeleton loading state
          <div className="flex items-center gap-4 animate-pulse">
            <div className="h-9 w-24 bg-gray-200 rounded-md dark:bg-zinc-800"></div>
            <div className="h-9 w-9 bg-gray-200 rounded-full dark:bg-zinc-800"></div>
            <div className="h-9 w-9 bg-gray-200 rounded-full dark:bg-zinc-800"></div>
          </div>
        ) : user ? (
          // Authenticated state
          <>
            {rightActions}
            <Link
              href="/posts/create"
              className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:bg-indigo-500 dark:hover:bg-indigo-600 transition duration-150"
            >
              Create Post
            </Link>
            <NotificationDropdown />
            <ProfileDropdown user={user} />
          </>
        ) : (
          // Unauthenticated/Guest state
          <Link
            href="/login"
            className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 transition duration-150 focus:outline-none dark:bg-indigo-500 dark:hover:bg-indigo-600"
          >
            Log In
          </Link>
        )}
      </div>
    </nav>
  );
}
