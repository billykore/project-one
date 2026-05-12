import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { Post } from "../model";
import { PostDetailState } from "./model";

export const usePostDetail = () => {
  const params = useParams();
  const router = useRouter();
  const id = params.id;
  
  const [state, setState] = useState<PostDetailState>({
    post: null,
    isLoading: true,
    error: null,
  });

  useEffect(() => {
    if (!id) return;

    const checkAuthAndFetchPost = async () => {
      try {
        // First verify authentication
        await api.get("/users/me");

        // Then fetch post
        const post = await api.get<Post>(`/posts/${id}`);
        setState({
          post,
          isLoading: false,
          error: null,
        });
      } catch (err) {
        if (err instanceof ApiError && err.status === 401) {
          router.push("/login");
          return;
        }

        const message = err instanceof Error ? err.message : "Failed to load post";
        setState({
          post: null,
          isLoading: false,
          error: message,
        });
      }
    };

    checkAuthAndFetchPost();
  }, [id, router]);

  return {
    ...state,
  };
};
