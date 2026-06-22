import {
  NavbarActionsSkeleton,
  PageSkeletonLayout,
  PostsGridSkeleton,
} from "@/components/ui/page-skeleton";

export default function PostsLoading() {
  return (
    <PageSkeletonLayout title="Posts" rightActions={<NavbarActionsSkeleton />}>
      <PostsGridSkeleton count={6} />
    </PageSkeletonLayout>
  );
}
