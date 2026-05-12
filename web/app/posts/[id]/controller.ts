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

    const fetchPost = async () => {
      try {
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

    fetchPost();
  }, [id, router]);

  return {
    ...state,
  };
};
