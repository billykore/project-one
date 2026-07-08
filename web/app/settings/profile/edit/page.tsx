import { redirect } from "next/navigation";
import Link from "next/link";
import { cookies } from "next/headers";
import { serverFetch } from "@/lib/server-fetch";
import { ApiError, handleApiResponse } from "@/lib/errors";
import { UserProfile } from "@/lib/types/profile.types";
import EditProfileForm from "@/components/profile/edit-profile-form";

function parseName(name: string): { firstName: string; lastName: string } {
  if (!name) return { firstName: "", lastName: "" };
  const parts = name.trim().split(/\s+/);
  if (parts.length === 1) return { firstName: parts[0], lastName: "" };
  return {
    firstName: parts[0],
    lastName: parts.slice(1).join(" "),
  };
}

export default async function EditProfilePage() {
  const cookieStore = await cookies();
  const username = cookieStore.get("username")?.value;

  if (!username) {
    redirect("/login");
  }

  let user: UserProfile;
  try {
    const res = await serverFetch(`/api/users/${username}`);
    user = await handleApiResponse<UserProfile>(res);
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 401) {
        redirect("/login");
      }
    }
    // For other errors, render a minimal error state.
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50 dark:bg-black">
        <div className="rounded-xl border border-zinc-200 bg-white p-8 text-center shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
          <p className="text-red-600 dark:text-red-400 font-medium">
            Failed to load profile data. Please try again.
          </p>
          <Link
            href="/"
            className="mt-4 inline-block text-sm font-semibold text-indigo-600 hover:underline dark:text-indigo-400"
          >
            Go to home
          </Link>
        </div>
      </div>
    );
  }

  const { firstName, lastName } = parseName(user.name);

  return (
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-black text-gray-900 dark:text-zinc-100 transition-colors duration-200">
      {/* Simple Navbar-like top bar */}
      <header className="border-b border-zinc-200 bg-white/80 backdrop-blur dark:border-zinc-800 dark:bg-zinc-900/80">
        <div className="mx-auto flex max-w-6xl items-center px-6 py-4">
          <a href={`/${username}`} className="text-sm font-medium text-zinc-500 hover:text-zinc-800 dark:text-zinc-400 dark:hover:text-zinc-200 transition-colors">
            &larr; Back to Profile
          </a>
          <h1 className="ml-4 text-lg font-bold text-zinc-900 dark:text-zinc-50">Edit Profile</h1>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex flex-1 items-start justify-center p-6 pt-10 md:p-8 md:pt-12">
        <EditProfileForm
          initialFirstName={firstName}
          initialLastName={lastName}
          initialUsername={user.username}
        />
      </main>
    </div>
  );
}
