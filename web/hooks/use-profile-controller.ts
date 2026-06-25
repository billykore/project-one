import { useState } from "react";
import { useRouter } from "next/navigation";
import { ApiError, handleApiResponse } from "@/lib/errors";
import { changePasswordSchema, ChangePasswordFormData, FollowerInfo } from "@/lib/types/profile.types";

export function useProfileController(profileUsername: string, initialFollowers: FollowerInfo[], initialFollowing: FollowerInfo[]) {
  const router = useRouter();
  // ponytail: read cookie directly in state initializer, no effect needed
  const [loggedInUsername] = useState<string | null>(() =>
    document.cookie.split("; ").find(r => r.startsWith("username="))?.split("=")[1] || null
  );

  const isOwner = loggedInUsername === profileUsername;

  // Follow states
  const [followers, setFollowers] = useState<FollowerInfo[]>(initialFollowers);
  const [following] = useState<FollowerInfo[]>(initialFollowing);
  const [isFollowingPending, setIsFollowingPending] = useState(false);

  const isFollowing = followers.some((f) => f.username === loggedInUsername);

  // Form states
  const [pwdForm, setPwdForm] = useState<ChangePasswordFormData>({
    oldPassword: "",
    newPassword: "",
    confirmPassword: "",
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isSubmittingPwd, setIsSubmittingPwd] = useState(false);
  const [pwdSuccess, setPwdSuccess] = useState<string | null>(null);

  const handleFollowToggle = async () => {
    if (!loggedInUsername) {
      router.push("/login");
      return;
    }
    setIsFollowingPending(true);
    try {
      if (isFollowing) {
        await handleApiResponse(await fetch(`/api/users/${profileUsername}/followers`, { method: "DELETE" }));
        setFollowers((prev) => prev.filter((f) => f.username !== loggedInUsername));
      } else {
        await handleApiResponse(await fetch(`/api/users/${profileUsername}/followers`, { method: "POST" }));
        setFollowers((prev) => [
          ...prev,
          {
            username: loggedInUsername,
            name: "You",
            followed_at: new Date().toISOString(),
            is_mutual: false,
          },
        ]);
      }
      router.refresh();
    } catch (err) {
      console.error("Failed to toggle follow status:", err);
    } finally {
      setIsFollowingPending(false);
    }
  };

  const handlePwdChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setPwdForm((prev) => ({ ...prev, [name]: value }));
    if (errors[name]) {
      setErrors((prev) => ({ ...prev, [name]: "" }));
    }
  };

  const submitPasswordChange = async (e: React.FormEvent) => {
    e.preventDefault();
    setPwdSuccess(null);

    // Client-side Zod validation
    const result = changePasswordSchema.safeParse(pwdForm);
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      result.error.issues.forEach((issue) => {
        const path = issue.path[0] as string;
        fieldErrors[path] = issue.message;
      });
      setErrors(fieldErrors);
      return;
    }

    setIsSubmittingPwd(true);
    try {
      await handleApiResponse(await fetch("/api/users/password", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          old_password: pwdForm.oldPassword,
          new_password: pwdForm.newPassword,
        }),
      }));
      setPwdSuccess("Password updated successfully!");
      setPwdForm({ oldPassword: "", newPassword: "", confirmPassword: "" });
      setErrors({});
    } catch (err) {
      if (err instanceof ApiError) {
        setErrors({ general: err.message });
      } else {
        setErrors({ general: "An unexpected error occurred" });
      }
    } finally {
      setIsSubmittingPwd(false);
    }
  };

  return {
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
  };
}
