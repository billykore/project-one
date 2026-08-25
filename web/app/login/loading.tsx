import { Skeleton } from "@/components/ui/skeleton";

export default function LoginLoading() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-paper px-4 py-12 sm:px-6">
      <div className="w-full max-w-sm space-y-5">
        <Skeleton className="h-3 w-24" />
        <Skeleton className="h-9 w-48" />
        <Skeleton className="h-4 w-full" />
        <div className="space-y-4 pt-4">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
        </div>
      </div>
    </div>
  );
}
