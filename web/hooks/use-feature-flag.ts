'use client';

import { useEffect, useState } from "react";
import { evaluateFeatureFlags, type FeatureFlagDecision } from "@/lib/feature-flags-api";

export function useFeatureFlag(key: string) {
  const [decision, setDecision] = useState<FeatureFlagDecision | null>(null);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    evaluateFeatureFlags([key])
      .then(([next]) => {
        if (active) setDecision(next ?? { key, enabled: false, source: "safe_default" });
      })
      .catch((err: unknown) => {
        if (active) {
          setError(err instanceof Error ? err : new Error("Unable to evaluate feature flag"));
          setDecision({ key, enabled: false, source: "safe_default" });
        }
      });
    return () => {
      active = false;
    };
  }, [key]);

  return {
    enabled: decision?.enabled ?? false,
    loading: decision === null && error === null,
    error,
    decision,
  };
}
