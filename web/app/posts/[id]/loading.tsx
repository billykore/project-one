import {
  PageSkeletonLayout,
  PostDetailContentSkeleton,
} from "@/components/ui/page-skeleton";
import { Skeleton } from "@/components/ui/skeleton";

export default function PostDetailLoading() {
  return (
    <PageSkeletonLayout
      title="Post Details"
      rightActions={<Skeleton className="h-9 w-28" />}
      contentMaxWidthClass="max-w-3xl"
    >
      <PostDetailContentSkeleton />
    </PageSkeletonLayout>
  );
}
