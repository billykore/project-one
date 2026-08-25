import React from "react";
import { Skeleton } from "@/components/ui/skeleton";
import SiteFooter from "@/components/layout/site-footer";

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
    <div className="flex min-h-screen flex-col bg-paper">
      <nav className="sticky top-0 z-40 border-b border-rule bg-paper/92 backdrop-blur-sm">
        <div className="mx-auto flex w-full max-w-6xl items-center gap-4 px-4 py-3 sm:px-6">
          <div className="flex items-baseline gap-2.5">
            <span className="font-body text-lg leading-none font-semibold tracking-tight text-ink">
              Project One
            </span>
            <span aria-hidden="true" className="meta text-faint">/</span>
            <span className="meta">{title}</span>
          </div>
          <div className="ml-auto flex items-center gap-1.5">{rightActions}</div>
        </div>
      </nav>

      <main className={`mx-auto w-full flex-1 px-4 py-8 sm:px-6 sm:py-10 ${contentMaxWidthClass}`}>
        {children}
      </main>

      <SiteFooter />
    </div>
  );
}

interface PostsGridSkeletonProps {
  count?: number;
}

export function PostsGridSkeleton({ count = 6 }: PostsGridSkeletonProps) {
  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: count }).map((_, index) => (
        <div key={index} className="entry flex flex-col p-5">
          <Skeleton className="h-3 w-24" />
          <Skeleton className="mt-4 h-6 w-3/4" />
          <div className="mt-4 space-y-2">
            <Skeleton className="h-3.5 w-full" />
            <Skeleton className="h-3.5 w-5/6" />
          </div>
          <Skeleton className="mt-5 h-3 w-20" />
        </div>
      ))}
    </div>
  );
}

export function PostDetailContentSkeleton() {
  return (
    <article className="panel p-6 sm:p-10">
      <header className="border-b border-rule pb-7">
        <Skeleton className="h-3 w-32" />
        <Skeleton className="mt-4 h-9 w-3/4" />
        <div className="mt-5 flex gap-3">
          <Skeleton className="h-3 w-16" />
          <Skeleton className="h-3 w-16" />
        </div>
      </header>

      <div className="mt-7 space-y-3">
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-5/6" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-4/5" />
      </div>
    </article>
  );
}

export function NavbarActionsSkeleton() {
  return (
    <div className="flex items-center gap-1.5">
      <Skeleton className="h-8 w-16" />
      <Skeleton className="h-8 w-8" />
      <Skeleton className="h-8 w-8" />
    </div>
  );
}
