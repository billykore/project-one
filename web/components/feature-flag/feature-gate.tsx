'use client';

import type { ReactNode } from "react";
import { useFeatureFlag } from "@/hooks/use-feature-flag";

type FeatureGateProps = {
  flag: string;
  children: ReactNode;
  fallback?: ReactNode;
};

export function FeatureGate({ flag, children, fallback = null }: FeatureGateProps) {
  const { enabled, loading } = useFeatureFlag(flag);
  if (loading || !enabled) return <>{fallback}</>;
  return <>{children}</>;
}
