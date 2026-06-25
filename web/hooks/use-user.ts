import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { ApiError, handleApiResponse } from "@/lib/errors";

interface User {
  username: string;
  email: string;
  name: string;
}

export function useUser() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const username = document.cookie.split("; ").find(r => r.startsWith("username="))?.split("=")[1];
    if (!username) {
      router.push("/login");
      return;
    }

    handleApiResponse<User>(await fetch(`/api/users/${username}`))
      .then(setUser)
      .catch((err) => {
        if (err instanceof ApiError && err.status === 401) {
          router.push("/login");
          return;
        }
        const msg = err instanceof Error ? err.message : "Failed to load user data";
        router.push(`/error?message=${encodeURIComponent(msg)}`);
      })
      .finally(() => setIsLoading(false));
  }, [router]);

  return { user, isLoading };
}
