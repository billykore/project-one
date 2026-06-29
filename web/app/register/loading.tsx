import { Skeleton } from "@/components/ui/skeleton";

export default function RegisterLoading() {
  return (
    <div className="min-h-screen grid lg:grid-cols-2 font-sans">
      {/* Left: Brand Panel Skeleton */}
      <div className="relative hidden lg:flex flex-col justify-between bg-linear-to-br from-indigo-600 via-indigo-700 to-purple-800 p-12 overflow-hidden">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,rgba(255,255,255,0.1)_0%,transparent_60%)]" />
        <div className="absolute top-0 right-0 w-96 h-96 bg-linear-to-bl from-purple-500/30 to-transparent rounded-full blur-3xl" />
        <div className="absolute bottom-0 left-0 w-80 h-80 bg-linear-to-tr from-indigo-400/20 to-transparent rounded-full blur-3xl" />

        <div className="relative z-10">
          <div className="flex items-center gap-3">
            <Skeleton className="h-10 w-10 rounded-xl bg-white/20" />
            <Skeleton className="h-5 w-24 bg-white/20" />
          </div>
        </div>

        <div className="relative z-10 space-y-3">
          <Skeleton className="h-9 w-56 bg-white/20" />
          <Skeleton className="h-9 w-48 bg-white/10" />
        </div>

        <div className="relative z-10">
          <Skeleton className="h-3 w-48 bg-white/10" />
        </div>
      </div>

      {/* Right: Form Panel Skeleton */}
      <div className="flex items-center justify-center px-4 py-12 sm:px-6 lg:px-12 bg-white dark:bg-gray-950">
        <div className="w-full max-w-md space-y-8">
          <div className="lg:hidden flex flex-col items-center gap-2">
            <Skeleton className="h-12 w-12 rounded-xl" />
            <Skeleton className="h-5 w-24" />
          </div>

          <div className="space-y-1">
            <Skeleton className="h-8 w-36" />
            <Skeleton className="h-4 w-64" />
          </div>

          <div className="space-y-5">
            <Skeleton className="h-10 w-full rounded-lg" />
            <Skeleton className="h-10 w-full rounded-lg" />
            <Skeleton className="h-10 w-full rounded-lg" />
            <Skeleton className="h-10 w-full rounded-lg" />
            <Skeleton className="h-10 w-full rounded-lg" />
            <Skeleton className="h-10 w-full rounded-lg" />
            <Skeleton className="h-10 w-full rounded-lg" />
          </div>

          <div className="flex justify-center">
            <Skeleton className="h-4 w-48" />
          </div>
        </div>
      </div>
    </div>
  );
}
