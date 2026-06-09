"use client";

import React, { useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { useProfileController } from "./controller";
import ResetPasswordForm from "./ResetPasswordForm";
import FollowListModal from "./FollowListModal";
import { UserProfile, FollowerInfo, PostInfo } from "./model";

interface UserProfileViewProps {
  profile: UserProfile;
  followers: FollowerInfo[];
  following: FollowerInfo[];
  posts: PostInfo[];
}

function getInitials(name: string): string {
  if (!name) return "?";
  const parts = name.trim().split(/\s+/);
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

function maskEmail(email: string): string {
  if (!email) return "";
  const parts = email.split("@");
  if (parts.length !== 2) return "••••••••";
  const [local, domain] = parts;

  const maskedLocal = local.length > 2
    ? local[0] + "•".repeat(local.length - 2) + local[local.length - 1]
    : "•".repeat(local.length);

  const domainParts = domain.split(".");
  if (domainParts.length < 2) {
    return `${maskedLocal}@••••.•••`;
  }

  const domainName = domainParts[0];
  const tld = domainParts.slice(1).join(".");

  const maskedDomain = domainName.length > 2
    ? domainName[0] + "•".repeat(domainName.length - 2) + domainName[domainName.length - 1]
    : "•".repeat(domainName.length);

  return `${maskedLocal}@${maskedDomain}.${tld}`;
}

export default function UserProfileView({
  profile,
  followers: initialFollowers,
  following: initialFollowing,
  posts,
}: UserProfileViewProps) {
  const {
    isOwner,
    isFollowing,
    isFollowingPending,
    handleFollowToggle,
    pwdForm,
    errors,
    isSubmittingPwd,
    pwdSuccess,
    handlePwdChange,
    submitPasswordChange,
    followers,
    following,
  } = useProfileController(profile.username, initialFollowers, initialFollowing);

  const [activeModal, setActiveModal] = useState<"followers" | "following" | null>(null);

  return (
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-black text-gray-900 dark:text-zinc-100 transition-colors duration-200">
      {/* Top Navbar */}
      <nav className="flex items-center justify-between bg-white px-6 py-4 shadow-sm dark:bg-zinc-900 border-b border-gray-100 dark:border-zinc-800">
        <div className="flex items-center gap-4">
          <Link href="/home">
            <Image
              className="dark:invert"
              src="/next.svg"
              alt="Next.js logo"
              width={90}
              height={18}
              priority
            />
          </Link>
          <span className="h-5 w-px bg-zinc-200 dark:bg-zinc-700"></span>
          <span className="text-md font-bold text-gray-900 dark:text-zinc-50">Profile</span>
        </div>
        <div className="flex items-center gap-3">
          <Link
            href="/posts"
            className="rounded-md border border-gray-300 bg-white px-3.5 py-1.5 text-xs font-semibold text-gray-700 hover:bg-gray-50 focus:outline-none dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-300 dark:hover:bg-zinc-700 transition duration-150"
          >
            All Posts
          </Link>
          <Link
            href="/home"
            className="rounded-md border border-gray-300 bg-white px-3.5 py-1.5 text-xs font-semibold text-gray-700 hover:bg-gray-50 focus:outline-none dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-300 dark:hover:bg-zinc-700 transition duration-150"
          >
            Dashboard
          </Link>
        </div>
      </nav>

      {/* Main Container */}
      <main className="mx-auto w-full max-w-6xl flex-1 p-6 md:p-8">
        {/* Profile Header Block */}
        <div className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 p-8 shadow-md dark:shadow-zinc-900/30 mb-8 text-white">
          <div className="absolute inset-0 bg-black/10 backdrop-blur-[1px]"></div>
          <div className="relative z-10 flex flex-col sm:flex-row items-center gap-6">
            <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-full bg-white/20 border-2 border-white/40 text-2xl font-bold uppercase tracking-wider text-white shadow-inner">
              {getInitials(profile.name)}
            </div>
            <div className="text-center sm:text-left flex-1">
              <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">{profile.name}</h1>
              <p className="text-sm font-medium text-white/80 mt-1">@{profile.username}</p>
            </div>
          </div>
        </div>

        {/* 2-Column Grid Layout */}
        <div className="grid gap-8 md:grid-cols-3">
          {/* Sidebar Area */}
          <div className="md:col-span-1 space-y-6">
            {/* User Meta Card */}
            <div className="rounded-xl border border-zinc-200 bg-white p-6 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
              <h3 className="text-sm font-semibold text-zinc-400 dark:text-zinc-500 uppercase tracking-wider mb-4">
                User Details
              </h3>
              <div className="space-y-4">
                <div>
                  <label className="text-xs text-zinc-400 dark:text-zinc-500 block font-medium">Email Address</label>
                  <p className="text-sm font-medium text-zinc-800 dark:text-zinc-200 mt-0.5 truncate" title={isOwner ? profile.email : maskEmail(profile.email)}>
                    {isOwner ? profile.email : maskEmail(profile.email)}
                  </p>
                </div>
                <div>
                  <label className="text-xs text-zinc-400 dark:text-zinc-500 block font-medium">Username</label>
                  <p className="text-sm font-medium text-zinc-800 dark:text-zinc-200 mt-0.5">
                    @{profile.username}
                  </p>
                </div>
              </div>

              {/* Followers & Following Stats */}
              <div className="flex gap-4 mt-6 pt-6 border-t border-zinc-100 dark:border-zinc-800">
                <button
                  onClick={() => setActiveModal("followers")}
                  className="flex-1 flex flex-col items-center p-3 rounded-lg border border-zinc-100 dark:border-zinc-800/80 hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition duration-150 cursor-pointer"
                  aria-label="View followers"
                >
                  <span className="text-lg font-bold text-zinc-900 dark:text-zinc-50">
                    {followers.length}
                  </span>
                  <span className="text-xs text-zinc-500 dark:text-zinc-400 mt-0.5">Followers</span>
                </button>

                <button
                  onClick={() => setActiveModal("following")}
                  className="flex-1 flex flex-col items-center p-3 rounded-lg border border-zinc-100 dark:border-zinc-800/80 hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition duration-150 cursor-pointer"
                  aria-label="View following list"
                >
                  <span className="text-lg font-bold text-zinc-900 dark:text-zinc-50">
                    {following.length}
                  </span>
                  <span className="text-xs text-zinc-500 dark:text-zinc-400 mt-0.5">Following</span>
                </button>
              </div>

              {/* Follow Toggle (if not profile owner) */}
              {!isOwner && (
                <div className="mt-6">
                  <button
                    onClick={handleFollowToggle}
                    disabled={isFollowingPending}
                    className={`w-full rounded-lg py-2.5 text-sm font-semibold shadow-sm transition duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 cursor-pointer ${
                      isFollowing
                        ? "bg-zinc-100 hover:bg-zinc-200 text-zinc-800 dark:bg-zinc-800 dark:hover:bg-zinc-700 dark:text-zinc-200 border border-zinc-200 dark:border-zinc-700"
                        : "bg-indigo-600 hover:bg-indigo-700 text-white dark:bg-indigo-500 dark:hover:bg-indigo-600"
                    } disabled:opacity-50`}
                  >
                    {isFollowingPending ? (
                      <span className="flex items-center justify-center gap-2">
                        <svg className="h-4 w-4 animate-spin text-current" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Loading...
                      </span>
                    ) : isFollowing ? (
                      "Unfollow"
                    ) : (
                      "Follow"
                    )}
                  </button>
                </div>
              )}
            </div>

            {/* Change Password Form (Owner only) */}
            {isOwner && (
              <ResetPasswordForm
                pwdForm={pwdForm}
                errors={errors}
                isSubmittingPwd={isSubmittingPwd}
                pwdSuccess={pwdSuccess}
                handlePwdChange={handlePwdChange}
                submitPasswordChange={submitPasswordChange}
              />
            )}
          </div>

          {/* Main Feed Content Area */}
          <div className="md:col-span-2 space-y-6">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-bold text-zinc-900 dark:text-zinc-50 flex items-center gap-2">
                <span>Posts</span>
                <span className="bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400 text-xs px-2.5 py-1 rounded-full font-semibold">
                  {posts.length}
                </span>
              </h2>
              {isOwner && (
                <Link
                  href="/posts/create"
                  className="text-xs font-semibold text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 hover:underline flex items-center gap-1"
                >
                  Create post +
                </Link>
              )}
            </div>

            {posts.length === 0 ? (
              <div className="flex flex-col items-center justify-center rounded-xl bg-white p-12 shadow-sm border border-zinc-200 dark:border-zinc-800 dark:bg-zinc-900 text-center">
                <svg className="w-10 h-10 text-zinc-300 dark:text-zinc-700 mb-3" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
                </svg>
                <p className="text-zinc-500 dark:text-zinc-400 mb-2 font-medium">No posts published yet.</p>
                {isOwner && (
                  <Link
                    href="/posts/create"
                    className="text-sm font-semibold text-indigo-600 dark:text-indigo-400 hover:underline"
                  >
                    Share your thoughts now
                  </Link>
                )}
              </div>
            ) : (
              <div className="grid gap-5 sm:grid-cols-2">
                {posts.map((post) => (
                  <Link
                    key={post.id}
                    href={`/posts/${post.id}`}
                    className="group flex flex-col justify-between rounded-xl bg-white p-5 shadow-sm border border-zinc-200 dark:border-zinc-800/80 dark:bg-zinc-900 hover:border-indigo-500/50 dark:hover:border-indigo-400/50 hover:shadow-md transition duration-200"
                  >
                    <div>
                      <h4 className="font-bold text-zinc-900 dark:text-zinc-50 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors line-clamp-1 mb-2">
                        {post.title}
                      </h4>
                      <p className="text-sm text-zinc-600 dark:text-zinc-400 line-clamp-3 mb-4">
                        {post.content}
                      </p>
                    </div>
                    <div className="flex flex-wrap items-center justify-between gap-2 pt-3 border-t border-zinc-100 dark:border-zinc-800/40 text-xs">
                      <span className="text-zinc-400">
                        {new Date(post.created_at).toLocaleDateString(undefined, {
                          year: 'numeric',
                          month: 'short',
                          day: 'numeric'
                        })}
                      </span>
                      {post.tags && post.tags.length > 0 && (
                        <div className="flex gap-1 overflow-hidden">
                          {post.tags.slice(0, 2).map((t) => (
                            <span key={t} className="bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400 px-2 py-0.5 rounded text-[10px] font-medium whitespace-nowrap">
                              #{t}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="bg-white py-6 border-t border-zinc-100 dark:border-zinc-800 text-center text-xs text-zinc-500 dark:bg-zinc-900 dark:text-zinc-400 mt-12">
        &copy; {new Date().getFullYear()} Project One. All rights reserved.
      </footer>

      {/* Modal overlays */}
      <FollowListModal
        isOpen={activeModal === "followers"}
        onClose={() => setActiveModal(null)}
        title="Followers"
        users={followers}
      />

      <FollowListModal
        isOpen={activeModal === "following"}
        onClose={() => setActiveModal(null)}
        title="Following"
        users={following}
      />
    </div>
  );
}
