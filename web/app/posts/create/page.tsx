import { Metadata } from "next";
import { CreatePostForm } from "@/components/posts/create-post-form";
import Navbar from "@/components/layout/navbar";

export const metadata: Metadata = {
  title: "Create Post | Project One",
  description: "Create a new post and share it with the world.",
};

export default function CreatePostPage() {
  return (
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-gray-950">
      <Navbar pageTitle="Create Post" />
      <main className="flex flex-1 items-center justify-center p-6 sm:p-8">
        <CreatePostForm />
      </main>
      <footer className="py-6 text-center text-xs text-gray-400 dark:text-gray-600">
        &copy; {new Date().getFullYear()} Project One. All rights reserved.
      </footer>
    </div>
  );
}
