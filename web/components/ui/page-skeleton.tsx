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
    <div className="flex min-h-screen flex-col bg-gray-50 font-sans dark:bg-gray-950">
      <nav className="sticky top-0 z-40 flex items-center justify-between bg-white/80 backdrop-blur-md px-6 py-3 shadow-sm ring-1 ring-gray-200 dark:bg-gray-950/80 dark:ring-gray-800">
        <div className="flex items-center gap-4">
          <Skeleton className="h-5 w-20" />
          <span className="text-sm font-medium text-gray-500 dark:text-gray-400">{title}</span>
        </div>
        <div className="flex items-center gap-4">{rightActions}</div>
      </nav>

      <main className={`mx-auto w-full flex-1 p-6 sm:p-8 ${contentMaxWidthClass}`}>{children}</main>

      <footer className="py-6 text-center text-xs text-gray-400 dark:text-gray-600">
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
        <div key={index} className="flex flex-col rounded-2xl bg-white p-6 shadow-sm ring-1 ring-gray-200 dark:bg-gray-900 dark:ring-gray-800">
          <Skeleton className="h-7 w-3/4" />
          <div className="mt-4 space-y-2">
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-5/6" />
          </div>
          <div className="mt-6 flex items-center justify-between">
            <Skeleton className="h-4 w-20" />
            <Skeleton className="h-4 w-16" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function PostDetailContentSkeleton() {
  return (
    <div className="rounded-2xl bg-white p-6 shadow-sm ring-1 ring-gray-200 dark:bg-gray-900 dark:ring-gray-800 sm:p-10">
      <header className="mb-8 border-b border-gray-100 pb-8 dark:border-gray-800">
        <Skeleton className="h-10 w-3/4" />
        <div className="mt-4 flex gap-4">
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-4 w-24" />
        </div>
        <div className="mt-5 flex gap-2">
          <Skeleton className="h-6 w-16 rounded-full" />
          <Skeleton className="h-6 w-16 rounded-full" />
        </div>
      </header>

      <div className="space-y-3">
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-5/6" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-4/5" />
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
