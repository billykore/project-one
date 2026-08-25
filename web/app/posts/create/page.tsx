import { Metadata } from "next";
import { CreatePostForm } from "@/components/posts/create-post-form";
import Navbar from "@/components/layout/navbar";
import SiteFooter from "@/components/layout/site-footer";

export const metadata: Metadata = {
  title: "Write a post | Project One",
  description: "Write something and publish it to your followers.",
};

export default function CreatePostPage() {
  return (
    <div className="flex min-h-screen flex-col bg-paper">
      <Navbar pageTitle="Write" />
      <main className="flex flex-1 items-start justify-center px-4 py-8 sm:px-6 sm:py-12">
        <CreatePostForm />
      </main>
      <SiteFooter />
    </div>
  );
}
