'use client';

import { useCallback, useEffect, useState } from "react";
import { archiveFeatureFlag, fetchFeatureFlagAdminList, fetchFeatureFlagAudit, setFeatureFlagEnvironment, setFeatureFlagOverrides, type FeatureFlagAuditEntry } from "@/lib/feature-flags-api";

type Setting = { environment: string; mode: string; rolloutPercentage: number; revision: number };
type Flag = { key: string; name: string; purpose: string; owner: string; lifecycle: string; safeDefault: boolean; settings?: Setting[] };

type FlagList = { flags: Flag[] };

export default function FeatureFlagsPage() {
  const [flags, setFlags] = useState<Flag[]>([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState<string | null>(null);
  const [audit, setAudit] = useState<{ key: string; items: FeatureFlagAuditEntry[] } | null>(null);
  const [overrideKey, setOverrideKey] = useState<string | null>(null);
  const [includeUsers, setIncludeUsers] = useState("");
  const [excludeUsers, setExcludeUsers] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const payload = (await fetchFeatureFlagAdminList()) as FlagList;
      setFlags(payload.flags ?? []);
      setMessage(null);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Unable to load feature flags");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  async function toggle(flag: Flag) {
    const setting = flag.settings?.find((item) => item.environment === "local");
    const enabled = setting?.mode === "enabled_all";
    try {
      await setFeatureFlagEnvironment(flag.key, "local", {
        mode: enabled ? "disabled_all" : "enabled_all",
        rolloutPercentage: enabled ? 0 : 100,
        revision: setting?.revision ?? 0,
        reason: enabled ? "Disable from operator dashboard" : "Enable from operator dashboard",
      });
      await load();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Unable to update feature flag");
    }
  }

  async function setRollout(flag: Flag, percentage: number) {
    const setting = flag.settings?.find((item) => item.environment === "local");
    try {
      await setFeatureFlagEnvironment(flag.key, "local", {
        mode: percentage === 0 ? "disabled_all" : percentage === 100 ? "enabled_all" : "gradual",
        rolloutPercentage: percentage,
        revision: setting?.revision ?? 0,
        reason: "Update rollout from operator dashboard",
      });
      await load();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Unable to update rollout");
    }
  }

  async function showAudit(key: string) {
    try {
      setAudit({ key, items: await fetchFeatureFlagAudit(key) });
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Unable to load audit history");
    }
  }

  async function saveOverrides(key: string) {
    try {
      await setFeatureFlagOverrides(key, {
        environment: "local",
        include: includeUsers.split(",").map((value) => value.trim()).filter(Boolean),
        exclude: excludeUsers.split(",").map((value) => value.trim()).filter(Boolean),
        reason: "Update user overrides from operator dashboard",
      });
      setOverrideKey(null);
      await load();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Unable to update overrides");
    }
  }

  async function archive(flag: Flag) {
    if (!window.confirm(`Archive ${flag.key}? This cannot be undone.`)) return;
    try {
      await archiveFeatureFlag(flag.key, "Archive from operator dashboard");
      await load();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Unable to archive feature flag");
    }
  }

  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-12">
      <div className="mb-10 flex items-end justify-between gap-4">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.2em] text-ink-muted">Operations</p>
          <h1 className="mt-2 font-serif text-4xl text-ink">Feature flags</h1>
        </div>
        <button className="rounded border border-ink/20 px-4 py-2 text-sm transition hover:bg-ink hover:text-paper" onClick={() => void load()}>
          Refresh
        </button>
      </div>
      {message && <p className="mb-5 rounded border border-red-300 bg-red-50 p-3 text-sm text-red-800">{message}</p>}
      {loading ? (
        <p className="text-ink-muted">Loading flags...</p>
      ) : flags.length === 0 ? (
        <p className="text-ink-muted">No feature flags configured.</p>
      ) : (
        <div className="space-y-3">
          {flags.map((flag) => {
            const local = flag.settings?.find((item) => item.environment === "local");
            const enabled = local?.mode === "enabled_all";
            return (
              <article key={flag.key} className="flex flex-wrap items-center justify-between gap-4 rounded-lg border border-ink/10 bg-white/60 p-5 shadow-sm">
                <div>
                  <h2 className="font-semibold text-ink">{flag.name}</h2>
                  <p className="mt-1 font-mono text-xs text-ink-muted">{flag.key} · {flag.owner}</p>
                  <p className="mt-2 text-sm text-ink-muted">{flag.purpose}</p>
                </div>
                <button
                  className={`rounded px-4 py-2 text-sm ${enabled ? "bg-ink text-paper" : "border border-ink/20"}`}
                  onClick={() => void toggle(flag)}
                  disabled={flag.lifecycle === "archived"}
                >
                  {enabled ? "Enabled locally" : "Disabled locally"}
                </button>
                <label className="flex items-center gap-2 text-sm">
                  <span className="text-ink-muted">Rollout</span>
                  <input
                    className="w-16 rounded border border-ink/20 bg-paper px-2 py-1"
                    type="number"
                    min="0"
                    max="100"
                    defaultValue={local?.rolloutPercentage ?? 0}
                    aria-label={`${flag.key} rollout percentage`}
                    onChange={(event) => void setRollout(flag, Number(event.target.value))}
                  />%
                </label>
                <button className="text-sm underline" onClick={() => void showAudit(flag.key)}>Audit</button>
                <button className="text-sm underline" onClick={() => setOverrideKey(flag.key)} disabled={flag.lifecycle === "archived"}>Overrides</button>
                <button className="text-sm text-red-700 underline" onClick={() => void archive(flag)} disabled={flag.lifecycle === "archived"}>Archive</button>
                {overrideKey === flag.key && (
                  <div className="basis-full border-t border-ink/10 pt-4">
                    <div className="grid gap-3 md:grid-cols-2">
                      <input className="rounded border border-ink/20 bg-paper px-3 py-2 text-sm" placeholder="Include usernames, comma-separated" value={includeUsers} onChange={(event) => setIncludeUsers(event.target.value)} />
                      <input className="rounded border border-ink/20 bg-paper px-3 py-2 text-sm" placeholder="Exclude usernames, comma-separated" value={excludeUsers} onChange={(event) => setExcludeUsers(event.target.value)} />
                    </div>
                    <div className="mt-3 flex gap-3">
                      <button className="rounded bg-ink px-3 py-2 text-sm text-paper" onClick={() => void saveOverrides(flag.key)}>Save overrides</button>
                      <button className="text-sm underline" onClick={() => setOverrideKey(null)}>Cancel</button>
                    </div>
                  </div>
                )}
              </article>
            );
          })}
        </div>
      )}
      {audit && (
        <section className="mt-8 border-t border-ink/10 pt-6">
          <h2 className="font-serif text-2xl text-ink">Audit: {audit.key}</h2>
          <ul className="mt-3 space-y-2 text-sm text-ink-muted">
            {audit.items.map((item, index) => <li key={`${item.createdAt}-${index}`}>{item.actor} - {item.reason} - {item.previousValue} -&gt; {item.newValue}</li>)}
          </ul>
        </section>
      )}
    </main>
  );
}
