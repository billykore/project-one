"use client";

import React, { useState, useRef, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { handleApiResponse } from "@/lib/errors";
import { ConfirmDialog } from "@/components/ui/modal";

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

const MENU = [
  { href: "/", label: "Home feed" },
  { href: "/posts", label: "All posts" },
  { href: "/posts/create", label: "Write a post" },
];

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

  // Escape closes the menu; the confirm dialog handles its own key events.
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") closeDropdown();
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
      await handleApiResponse(await fetch("/api/logout", { method: "POST" }));
    } catch (err) {
      console.error("Logout failed:", err);
    } finally {
      localStorage.removeItem("username");
      setIsLogoutModalOpen(false);
      setIsLoggingOut(false);
      router.push("/login");
    }
  };

  return (
    <div className="relative" ref={containerRef}>
      <button
        onClick={toggleDropdown}
        className="mark h-8 w-8 text-[0.65rem]"
        aria-expanded={isOpen}
        aria-haspopup="true"
        title="Account Menu"
      >
        {getInitials(user.name)}
      </button>

      {isOpen && (
        <div className="pop enter-pop absolute right-0 z-50 mt-2 w-60 overflow-hidden">
          <div className="border-b border-rule px-4 py-3">
            <p className="meta">Signed in as</p>
            <p className="mt-1 truncate text-sm font-semibold text-ink">{user.name}</p>
            <p className="meta truncate normal-case">@{user.username}</p>
          </div>

          <div className="py-1">
            {MENU.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                onClick={closeDropdown}
                className="block px-4 py-2 text-sm text-ink transition-colors hover:bg-sunken"
              >
                {item.label}
              </Link>
            ))}
            <Link
              href={`/${user.username}`}
              onClick={closeDropdown}
              className="block px-4 py-2 text-sm text-ink transition-colors hover:bg-sunken"
            >
              Your profile
            </Link>
          </div>

          <div className="border-t border-rule py-1">
            <button
              onClick={handleLogoutClick}
              className="block w-full cursor-pointer px-4 py-2 text-left text-sm font-medium text-ember transition-colors hover:bg-ember-soft"
            >
              Log out
            </button>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={isLogoutModalOpen}
        onClose={() => setIsLogoutModalOpen(false)}
        onConfirm={handleConfirmLogout}
        rubric="End session"
        title="Log out of Project One?"
        description="You will need your email and password to get back in."
        confirmLabel="Log out"
        pendingLabel="Logging out"
        pending={isLoggingOut}
      />
    </div>
  );
}
