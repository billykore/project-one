import { notFound, redirect } from "next/navigation";
import { Suspense } from "react";
import { serverFetch } from "@/lib/server-fetch";
import { ApiError, handleApiResponse } from "@/lib/errors";
import UserProfileView from "@/components/profile/user-profile-view";
import { UserProfile, FollowerInfo, PostInfo } from "@/lib/types/profile.types";

interface PageProps {
  params: Promise<{ username: string }>;
}

type CursorPage<T> = {
  data: T[];
  next_cursor: string;
  has_more: boolean;
};

export default async function UserProfilePage({ params }: PageProps) {
  const { username } = await params;

  let user: UserProfile;
  let followers: FollowerInfo[];
  let following: FollowerInfo[];
  let posts: PostInfo[];

  try {
    // Parallel fetch to optimize response time & prevent waterfalls
    const [userData, followersData, followingData, postsData] = await Promise.all([
      serverFetch(`/api/users/${username}`).then(r => handleApiResponse<UserProfile>(r)),
      serverFetch(`/api/users/${username}/followers`).then(r => handleApiResponse<CursorPage<FollowerInfo>>(r)),
      serverFetch(`/api/users/${username}/following`).then(r => handleApiResponse<CursorPage<FollowerInfo>>(r)),
      serverFetch(`/api/users/${username}/posts`).then(r => handleApiResponse<CursorPage<PostInfo>>(r)),
    ]);
    user = userData;
    followers = followersData.data || [];
    following = followingData.data || [];
    posts = postsData.data || [];
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 401) {
        redirect("/login");
      }
      if (err.status === 404) {
        notFound();
      }
    }
    throw err;
  }

  return (
    <Suspense fallback={null}>
      <UserProfileView
        key={user.username}
        profile={user}
        followers={followers}
        following={following}
        posts={posts}
      />
    </Suspense>
  );
}
