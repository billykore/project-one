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
      <div className="flex min-h-screen items-center justify-center bg-paper p-4">
        <div className="panel measure border-t-2 border-t-ember p-8">
          <p className="meta">Profile unavailable</p>
          <h1 className="mt-2 text-2xl font-semibold text-ink">We couldn&apos;t load your details</h1>
          <p className="prose-lead mt-3">The server didn&apos;t answer. Try again in a moment.</p>
          <Link href="/" className="btn btn-primary mt-6">
            Go home
          </Link>
        </div>
      </div>
    );
  }

  const { firstName, lastName } = parseName(user.name);

  return (
    <div className="flex min-h-screen flex-col bg-paper">
      <header className="border-b border-rule bg-paper/92 backdrop-blur-sm">
        <div className="mx-auto flex w-full max-w-6xl items-baseline gap-2.5 px-4 py-3 sm:px-6">
          <span className="font-body text-lg leading-none font-semibold tracking-tight text-ink">
            Project One
          </span>
          <span aria-hidden="true" className="meta text-faint">/</span>
          <span className="meta">Edit profile</span>
          <Link href={`/${username}`} className="meta ml-auto hover:text-accent">
            Back to profile
          </Link>
        </div>
      </header>

      <main className="flex flex-1 items-start justify-center px-4 py-10 sm:px-6 sm:py-14">
        <EditProfileForm
          initialFirstName={firstName}
          initialLastName={lastName}
          initialUsername={user.username}
        />
      </main>
    </div>
  );
}
