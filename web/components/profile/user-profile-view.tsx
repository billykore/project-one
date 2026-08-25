"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useProfileController } from "@/hooks/use-profile-controller";
import ResetPasswordForm from "@/components/profile/reset-password-form";
import FollowListModal from "@/components/profile/follow-list-modal";
import { UserProfile, FollowerInfo, PostInfo } from "@/lib/types/profile.types";
import Navbar from "@/components/layout/navbar";
import SiteFooter from "@/components/layout/site-footer";
import { Spinner } from "@/components/ui/modal";

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
  const searchParams = useSearchParams();
  const [successBanner, setSuccessBanner] = useState<string | null>(() => {
    return searchParams.get("updated") === "1" ? "Your profile is updated." : null;
  });

  // Auto-dismiss success banner after 5 seconds.
  useEffect(() => {
    if (successBanner) {
      const timer = setTimeout(() => setSuccessBanner(null), 5000);
      return () => clearTimeout(timer);
    }
  }, [successBanner]);

  const displayEmail = isOwner ? profile.email : maskEmail(profile.email);

  return (
    <div className="flex min-h-screen flex-col bg-paper">
      <Navbar pageTitle="Profile" />

      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:px-6 sm:py-10">
        {successBanner && (
          <div className="notice notice-good enter mb-8 items-center justify-between">
            <span>{successBanner}</span>
            <button
              onClick={() => setSuccessBanner(null)}
              className="btn btn-ghost btn-sm -my-1"
              aria-label="Dismiss success message"
            >
              Dismiss
            </button>
          </div>
        )}

        {/* Colophon: who this is, in the app's own margin voice. */}
        <header className="flex flex-wrap items-center gap-5 border-b border-rule pb-8">
          <span aria-hidden="true" className="mark h-16 w-16 text-lg">
            {getInitials(profile.name)}
          </span>
          <div className="min-w-0 flex-1">
            <h1 className="text-3xl font-semibold text-ink sm:text-4xl">{profile.name}</h1>
            <p className="meta-lg mt-1.5">@{profile.username}</p>
          </div>
          {!isOwner && (
            <button
              onClick={handleFollowToggle}
              disabled={isFollowingPending}
              className={`btn btn-lg ${isFollowing ? "btn-quiet" : "btn-primary"}`}
            >
              {isFollowingPending && <Spinner />}
              {isFollowing ? "Unfollow" : "Follow"}
            </button>
          )}
        </header>

        <div className="mt-8 grid gap-8 md:grid-cols-3">
          <div className="space-y-6 md:col-span-1">
            <section className="panel p-6">
              <h2 className="rubric">Details</h2>
              <dl className="mt-4 space-y-4">
                <div>
                  <dt className="meta">Email</dt>
                  <dd className="mt-1 truncate font-mono text-sm text-ink" title={displayEmail}>
                    {displayEmail}
                  </dd>
                </div>
                <div>
                  <dt className="meta">Username</dt>
                  <dd className="mt-1 font-mono text-sm text-ink">@{profile.username}</dd>
                </div>
              </dl>

              <div className="mt-6 grid grid-cols-2 gap-2 border-t border-rule pt-6">
                <button
                  onClick={() => setActiveModal("followers")}
                  className="rounded-sm border border-rule px-3 py-3 text-left transition-colors hover:border-rule-strong hover:bg-sunken"
                  aria-label="View followers"
                >
                  <span className="block font-mono text-xl font-semibold tabular-nums text-ink">
                    {followers.length}
                  </span>
                  <span className="meta">Followers</span>
                </button>
                <button
                  onClick={() => setActiveModal("following")}
                  className="rounded-sm border border-rule px-3 py-3 text-left transition-colors hover:border-rule-strong hover:bg-sunken"
                  aria-label="View following list"
                >
                  <span className="block font-mono text-xl font-semibold tabular-nums text-ink">
                    {following.length}
                  </span>
                  <span className="meta">Following</span>
                </button>
              </div>

              {isOwner && (
                <Link href="/settings/profile/edit" className="btn btn-quiet btn-block mt-4">
                  Edit profile
                </Link>
              )}
            </section>

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

          <div className="space-y-4 md:col-span-2">
            <h2 className="rubric">
              {posts.length === 1 ? "1 post" : `${posts.length} posts`}
            </h2>

            {posts.length === 0 ? (
              <div className="panel px-6 py-14 text-center">
                <p className="prose-lead">
                  {isOwner ? "You haven't published anything yet." : `${profile.name} hasn't published anything yet.`}
                </p>
                {isOwner && (
                  <Link href="/posts/create" className="btn btn-primary mt-5">
                    Write your first post
                  </Link>
                )}
              </div>
            ) : (
              <div className="grid gap-3 sm:grid-cols-2">
                {posts.map((post) => (
                  <Link
                    key={post.id}
                    href={`/posts/${post.id}`}
                    className="entry flex flex-col p-5"
                  >
                    <time dateTime={post.created_at} className="meta">
                      {new Date(post.created_at).toLocaleDateString(undefined, {
                        day: "numeric",
                        month: "short",
                        year: "numeric",
                      })}
                    </time>
                    <h3 className="mt-2.5 line-clamp-2 text-lg leading-snug font-semibold text-ink">
                      {post.title}
                    </h3>
                    <p className="prose-lead mt-2 line-clamp-3 text-sm">{post.content}</p>
                    {post.tags && post.tags.length > 0 && (
                      <div className="mt-auto flex flex-wrap gap-x-2.5 gap-y-1 pt-4">
                        {post.tags.slice(0, 3).map((t) => (
                          <span key={t} className="tag">{t}</span>
                        ))}
                      </div>
                    )}
                  </Link>
                ))}
              </div>
            )}
          </div>
        </div>
      </main>

      <SiteFooter />

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
