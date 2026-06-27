import { Skeleton } from "@/components/ui/skeleton";

export default function LoginLoading() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-white px-4 py-12 dark:bg-gray-950 sm:px-6 lg:px-12">
      <div className="w-full max-w-md space-y-5">
        {/* Mobile logo */}
        <div className="flex flex-col items-center gap-2 lg:hidden">
          <Skeleton className="h-12 w-12 rounded-xl" />
          <Skeleton className="h-5 w-24" />
        </div>

        {/* Form heading */}
        <div className="space-y-1">
          <Skeleton className="h-8 w-24" />
          <Skeleton className="h-4 w-64" />
        </div>

        {/* Email field */}
        <Skeleton className="h-10 w-full rounded-lg" />

        {/* Password field */}
        <Skeleton className="h-10 w-full rounded-lg" />

        {/* Links row */}
        <div className="flex items-center justify-between pt-1">
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-4 w-28" />
        </div>

        {/* Submit button */}
        <Skeleton className="h-10 w-full rounded-lg" />
      </div>
    </div>
  );
}
