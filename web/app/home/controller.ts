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
  const [isLogoutModalOpen, setIsLogoutModalOpen] = useState(false);
  const [isLoggingOut, setIsLoggingOut] = useState(false);

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

        const message = err instanceof Error ? err.message : "Failed to load user data";
        router.push(`/error?message=${encodeURIComponent(message)}`);
      }
    };

    fetchUser();
  }, [router]);

  const openLogoutModal = () => setIsLogoutModalOpen(true);
  const closeLogoutModal = () => setIsLogoutModalOpen(false);

  const handleLogout = async () => {
    setIsLoggingOut(true);
    try {
      await api.post("/users/logout", {});
      router.push("/login");
    } catch (err) {
      console.error("Logout failed:", err);
      // Even if logout fails on server, we might want to redirect
      router.push("/login");
    } finally {
      setIsLoggingOut(false);
      closeLogoutModal();
    }
  };

  return {
    ...state,
    isLogoutModalOpen,
    isLoggingOut,
    openLogoutModal,
    closeLogoutModal,
    handleLogout,
  };
};
