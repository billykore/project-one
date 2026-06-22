import React from "react";
import { Skeleton } from "@/components/ui/skeleton";

interface PageSkeletonLayoutProps {
  title: string;
  children: React.ReactNode;
  rightActions?: React.ReactNode;
  contentMaxWidthClass?: string;
}

export function PageSkeletonLayout({
  title,
  children,
  rightActions,
  contentMaxWidthClass = "max-w-5xl",
}: PageSkeletonLayoutProps) {
  return (
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-black">
      <nav className="flex items-center justify-between border-b border-gray-100 bg-white px-8 py-4 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
        <div className="flex items-center gap-4">
          <Skeleton className="h-5 w-24" />
          <span className="text-xl font-bold text-gray-900 dark:text-zinc-50">{title}</span>
        </div>
        <div className="flex items-center gap-4">{rightActions}</div>
      </nav>

      <main className={`mx-auto w-full flex-1 p-8 ${contentMaxWidthClass}`}>{children}</main>

      <footer className="bg-white py-6 text-center text-sm text-gray-500 dark:bg-zinc-900 dark:text-zinc-400">
        &copy; {new Date().getFullYear()} Project One. All rights reserved.
      </footer>
    </div>
  );
}

interface PostsGridSkeletonProps {
  count?: number;
}

export function PostsGridSkeleton({ count = 6 }: PostsGridSkeletonProps) {
  return (
    <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: count }).map((_, index) => (
        <div key={index} className="flex flex-col rounded-xl bg-white p-6 shadow-md dark:bg-zinc-900">
          <Skeleton className="h-7 w-3/4" />
          <div className="mt-4 space-y-2">
            <Skeleton className="h-4 w-full bg-gray-100 dark:bg-zinc-800" />
            <Skeleton className="h-4 w-5/6 bg-gray-100 dark:bg-zinc-800" />
          </div>
          <div className="mt-6 flex items-center justify-between">
            <Skeleton className="h-4 w-20 bg-gray-100 dark:bg-zinc-800" />
            <Skeleton className="h-4 w-16 bg-indigo-100 dark:bg-indigo-950" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function PostDetailContentSkeleton() {
  return (
    <div className="rounded-xl bg-white p-8 shadow-lg dark:bg-zinc-900 sm:p-12">
      <header className="mb-8 border-b border-gray-100 pb-8 dark:border-zinc-800">
        <Skeleton className="h-10 w-3/4" />
        <div className="mt-4 flex gap-4">
          <Skeleton className="h-4 w-32 bg-gray-100 dark:bg-zinc-800" />
        </div>
        <div className="mt-6 flex gap-2">
          <Skeleton className="h-6 w-16 rounded-full bg-indigo-50 dark:bg-indigo-950" />
          <Skeleton className="h-6 w-16 rounded-full bg-indigo-50 dark:bg-indigo-950" />
        </div>
      </header>

      <div className="space-y-4">
        <Skeleton className="h-4 w-full bg-gray-100 dark:bg-zinc-800" />
        <Skeleton className="h-4 w-full bg-gray-100 dark:bg-zinc-800" />
        <Skeleton className="h-4 w-5/6 bg-gray-100 dark:bg-zinc-800" />
        <Skeleton className="h-4 w-full bg-gray-100 dark:bg-zinc-800" />
        <Skeleton className="h-4 w-4/5 bg-gray-100 dark:bg-zinc-800" />
      </div>
    </div>
  );
}

export function NavbarActionsSkeleton() {
  return (
    <div className="flex items-center gap-4">
      <Skeleton className="h-9 w-24" />
      <Skeleton className="h-9 w-9 rounded-full" />
      <Skeleton className="h-9 w-9 rounded-full" />
    </div>
  );
}
