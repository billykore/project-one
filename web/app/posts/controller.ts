import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { Post, PostsState } from "./model";

export const usePosts = () => {
  const router = useRouter();
  const [state, setState] = useState<PostsState>({
    posts: [],
    isLoading: true,
    error: null,
  });

  useEffect(() => {
    const checkAuthAndFetchPosts = async () => {
      try {
        // First verify authentication
        await api.get("/users/me");
        
        // Then fetch posts
        const posts = await api.get<Post[]>("/posts");
        setState({
          posts: posts || [],
          isLoading: false,
          error: null,
        });
      } catch (err) {
        // If 401 Unauthorized, redirect to login
        if (err instanceof ApiError && err.status === 401) {
          router.push("/login");
          return;
        }

        const message = err instanceof Error ? err.message : "Failed to load posts";
        setState(prev => ({
          ...prev,
          isLoading: false,
          error: message,
        }));
        router.push(`/error?message=${encodeURIComponent(message)}`);
      }
    };

    checkAuthAndFetchPosts();
  }, [router]);

  const truncateContent = (content: string, limit: number = 100) => {
    if (content.length <= limit) return content;
    return content.slice(0, limit) + "...";
  };

  return {
    ...state,
    truncateContent,
  };
};
