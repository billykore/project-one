"use client";

import React, { useEffect } from "react";
import Link from "next/link";
import { FollowerInfo } from "@/lib/types/profile.types";

interface FollowListModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  users: FollowerInfo[];
}

export default function FollowListModal({ isOpen, onClose, title, users }: FollowListModalProps) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isOpen) {
        onClose();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm transition-opacity"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
    >
      <div 
        className="w-full max-w-sm rounded-xl bg-white p-6 shadow-xl dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 transform transition-all"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex justify-between items-center mb-4">
          <h3 id="modal-title" className="text-lg font-bold text-gray-900 dark:text-zinc-50">{title}</h3>
          <button 
            onClick={onClose} 
            className="text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-300 cursor-pointer text-lg font-bold"
            aria-label="Close modal"
          >
            ✕
          </button>
        </div>
        <div className="max-h-60 overflow-y-auto divide-y divide-zinc-100 dark:divide-zinc-800 pr-1">
          {users.length === 0 ? (
            <p className="py-8 text-center text-sm text-zinc-500 dark:text-zinc-400">No users found.</p>
          ) : (
            users.map((u) => (
              <div key={u.username} className="flex justify-between items-center py-3">
                <div>
                  <Link
                    href={`/${u.username}`}
                    onClick={onClose}
                    className="font-medium text-indigo-600 hover:underline dark:text-indigo-400 transition-colors"
                  >
                    {u.name}
                  </Link>
                  <p className="text-xs text-zinc-500 dark:text-zinc-400">@{u.username}</p>
                </div>
                {u.is_mutual && (
                  <span className="text-[10px] bg-zinc-100 text-zinc-600 px-2 py-0.5 rounded-full dark:bg-zinc-800 dark:text-zinc-400 font-medium">
                    Mutual
                  </span>
                )}
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
