import { Metadata } from "next";
import { CreatePostForm } from "./CreatePostForm";

export const metadata: Metadata = {
  title: "Create Post | Project One",
  description: "Create a new post and share it with the world.",
};

export default function CreatePostPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <CreatePostForm />
    </div>
  );
}
