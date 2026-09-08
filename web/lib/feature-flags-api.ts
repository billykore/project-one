import { handleApiResponse } from "./errors";

export type FeatureFlagDecision = {
  key: string;
  enabled: boolean;
  source: string;
};

type FeatureFlagEvaluateResponse = {
  environment: string;
  decisions: FeatureFlagDecision[];
};

export async function evaluateFeatureFlags(keys: string[]): Promise<FeatureFlagDecision[]> {
  if (keys.length === 0) return [];
  const params = new URLSearchParams({ keys: keys.join(",") });
  const response = await fetch(`/api/feature-flags?${params.toString()}`);
  const payload = await handleApiResponse<FeatureFlagEvaluateResponse>(response);
  return payload.decisions;
}

export async function fetchFeatureFlagAdminList(): Promise<unknown> {
  return handleApiResponse(await fetch("/api/admin/feature-flags"));
}

export async function setFeatureFlagEnvironment(
  key: string,
  environment: string,
  body: { mode: string; rolloutPercentage: number; revision: number; reason: string },
): Promise<void> {
  await handleApiResponse(
    await fetch(`/api/admin/feature-flags/${encodeURIComponent(key)}/environment/${encodeURIComponent(environment)}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }),
  );
}

export type FeatureFlagAuditEntry = {
  field: string;
  environment?: string;
  previousValue: string;
  newValue: string;
  actor: string;
  reason: string;
  createdAt: string;
};

export async function fetchFeatureFlagAudit(key: string): Promise<FeatureFlagAuditEntry[]> {
  const response = await fetch(`/api/admin/feature-flags/${encodeURIComponent(key)}/audit`);
  const payload = await handleApiResponse<{ items?: FeatureFlagAuditEntry[] }>(response);
  return payload.items ?? [];
}

export async function setFeatureFlagOverrides(
  key: string,
  body: { environment: string; include: string[]; exclude: string[]; reason: string },
): Promise<void> {
  await handleApiResponse(
    await fetch(`/api/admin/feature-flags/${encodeURIComponent(key)}/overrides`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }),
  );
}

export async function archiveFeatureFlag(key: string, reason: string): Promise<void> {
  await handleApiResponse(
    await fetch(`/api/admin/feature-flags/${encodeURIComponent(key)}/archive`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ reason }),
    }),
  );
}
