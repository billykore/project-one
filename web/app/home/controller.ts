import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { UserResponse, HomeState } from "./model";

export const useHome = () => {
  const router = useRouter();
  const [state, setState] = useState<HomeState>({
    user: null,
    isLoading: true,
    error: null,
  });

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const userData = await api.get<UserResponse>("/users/me");
        setState({
          user: userData,
          isLoading: false,
          error: null,
        });
      } catch (err) {
        // If 401 Unauthorized, redirect to login
        if (err instanceof ApiError && err.status === 401) {
          router.push("/login");
          return;
        }

        setState({
          user: null,
          isLoading: false,
          error: err instanceof Error ? err.message : "Failed to load user data",
        });
      }
    };

    fetchUser();
  }, [router]);

  const handleLogout = async () => {
    try {
      await api.post("/users/logout", {});
      router.push("/login");
    } catch (err) {
      console.error("Logout failed:", err);
      // Even if logout fails on server, we might want to redirect
      router.push("/login");
    }
  };

  return {
    ...state,
    handleLogout,
  };
};
