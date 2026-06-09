import { notFound, redirect } from "next/navigation";
import { apiServer } from "@/lib/api-server";
import { ApiError } from "@/lib/api";
import UserProfileView from "./UserProfileView";
import { UserProfile, FollowerInfo, PostInfo } from "./model";

interface PageProps {
  params: Promise<{ username: string }>;
}

export default async function UserProfilePage({ params }: PageProps) {
  const { username } = await params;

  let user: UserProfile;
  let followers: FollowerInfo[];
  let following: FollowerInfo[];
  let posts: PostInfo[];

  try {
    // Parallel fetch to optimize response time & prevent waterfalls
    const [userData, followersData, followingData, postsData] = await Promise.all([
      apiServer.get<UserProfile>(`/api/v1/users/${username}`),
      apiServer.get<FollowerInfo[]>(`/api/v1/users/${username}/followers`),
      apiServer.get<FollowerInfo[]>(`/api/v1/users/${username}/following`),
      apiServer.get<PostInfo[]>(`/api/v1/users/${username}/posts`),
    ]);
    user = userData;
    followers = followersData || [];
    following = followingData || [];
    posts = postsData || [];
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
    <UserProfileView
      key={user.username}
      profile={user}
      followers={followers}
      following={following}
      posts={posts}
    />
  );
}
