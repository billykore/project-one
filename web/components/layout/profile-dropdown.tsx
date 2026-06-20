"use client";

import React, { useState, useRef, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";

interface User {
  name: string;
  username: string;
  email: string;
}

interface ProfileDropdownProps {
  user: User;
}

function getInitials(name: string): string {
  if (!name) return "?";
  const parts = name.trim().split(/\s+/);
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

export default function ProfileDropdown({ user }: ProfileDropdownProps) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [isLogoutModalOpen, setIsLogoutModalOpen] = useState(false);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const containerRef = useRef<HTMLDivElement | null>(null);

  const toggleDropdown = () => setIsOpen((prev) => !prev);
  const closeDropdown = () => setIsOpen(false);

  // Close dropdown on click outside
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (!containerRef.current) return;
      if (e.target instanceof Node && !containerRef.current.contains(e.target)) {
        closeDropdown();
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Keyboard navigation for escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        closeDropdown();
        setIsLogoutModalOpen(false);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  const handleLogoutClick = () => {
    closeDropdown();
    setIsLogoutModalOpen(true);
  };

  const handleConfirmLogout = async () => {
    setIsLoggingOut(true);
    try {
      await api.post("/api/v1/auth/logout", {});
      localStorage.removeItem("username");
      setIsLogoutModalOpen(false);
      router.push("/login");
    } catch (err) {
      console.error("Logout failed:", err);
      localStorage.removeItem("username");
      setIsLogoutModalOpen(false);
      router.push("/login");
    } finally {
      setIsLoggingOut(false);
    }
  };

  return (
    <div className="relative" ref={containerRef}>
      {/* Profile Avatar Button */}
      <button
        onClick={toggleDropdown}
        className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 text-xs font-extrabold uppercase tracking-wider text-white shadow-sm hover:opacity-90 hover:scale-105 active:scale-95 transition-all duration-150 border border-white/20 dark:border-zinc-800 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 cursor-pointer"
        aria-expanded={isOpen}
        aria-haspopup="true"
        title="Account Menu"
      >
        {getInitials(user.name)}
      </button>

      {/* Dropdown Menu */}
      {isOpen && (
        <div className="absolute right-0 mt-2 w-56 rounded-xl border border-gray-100 bg-white/95 py-2 shadow-xl backdrop-blur-md dark:border-zinc-800 dark:bg-zinc-950/95 z-50 animate-in fade-in slide-in-from-top-1 duration-150">
          {/* User Info Header */}
          <div className="px-4 py-2 border-b border-gray-50 dark:border-zinc-800/50">
            <p className="text-xs text-gray-400 dark:text-zinc-500 font-medium">Signed in as</p>
            <p className="text-sm font-semibold text-gray-800 dark:text-zinc-200 truncate">{user.name}</p>
            <p className="text-xs text-gray-500 dark:text-zinc-400 truncate">@{user.username}</p>
          </div>

          {/* Menu Items */}
          <div className="py-1">
            <Link
              href="/home"
              onClick={closeDropdown}
              className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 dark:text-zinc-300 dark:hover:bg-zinc-900 transition-colors"
            >
              <svg className="mr-3 h-4 w-4 text-gray-400 dark:text-zinc-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
              </svg>
              Dashboard
            </Link>

            <Link
              href={`/${user.username}`}
              onClick={closeDropdown}
              className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 dark:text-zinc-300 dark:hover:bg-zinc-900 transition-colors"
            >
              <svg className="mr-3 h-4 w-4 text-gray-400 dark:text-zinc-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
              </svg>
              View Profile
            </Link>

            <Link
              href="/posts"
              onClick={closeDropdown}
              className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 dark:text-zinc-300 dark:hover:bg-zinc-900 transition-colors"
            >
              <svg className="mr-3 h-4 w-4 text-gray-400 dark:text-zinc-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round" d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 4a2 2 0 00-2-2m2 2v10a2 2 0 01-2 2M9 9h6m-6 4h6m-6 4h5" />
              </svg>
              All Posts
            </Link>

            <Link
              href="/posts/create"
              onClick={closeDropdown}
              className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 dark:text-zinc-300 dark:hover:bg-zinc-900 transition-colors"
            >
              <svg className="mr-3 h-4 w-4 text-gray-400 dark:text-zinc-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
              </svg>
              Create New Post
            </Link>
          </div>

          {/* Divider */}
          <div className="border-t border-gray-50 dark:border-zinc-800/50 my-1"></div>

          {/* Logout Option */}
          <div className="py-1">
            <button
              onClick={handleLogoutClick}
              className="flex w-full items-center px-4 py-2 text-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-950/30 transition-colors text-left cursor-pointer font-medium"
            >
              <svg className="mr-3 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
              </svg>
              Log Out
            </button>
          </div>
        </div>
      )}

      {/* Logout Confirmation Modal */}
      {isLogoutModalOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm animate-in fade-in duration-200"
          onClick={() => setIsLogoutModalOpen(false)}
          role="dialog"
          aria-modal="true"
          aria-labelledby="modal-title"
          aria-describedby="modal-desc"
        >
          <div
            className="w-full max-w-sm rounded-xl bg-white p-6 shadow-xl dark:bg-zinc-900 border border-gray-100 dark:border-zinc-800 animate-in zoom-in-95 duration-200"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 id="modal-title" className="text-xl font-bold text-gray-900 dark:text-zinc-50">
              Confirm Logout
            </h3>
            <p id="modal-desc" className="mt-4 text-gray-600 dark:text-zinc-400 text-sm">
              Are you sure you want to log out? You will need to log back in to access your account.
            </p>
            <div className="mt-8 flex justify-end gap-3">
              <button
                onClick={() => setIsLogoutModalOpen(false)}
                disabled={isLoggingOut}
                className="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none disabled:opacity-50 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-300 dark:hover:bg-zinc-700 cursor-pointer"
              >
                Cancel
              </button>
              <button
                onClick={handleConfirmLogout}
                disabled={isLoggingOut}
                className="inline-flex items-center rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 disabled:opacity-50 cursor-pointer"
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
