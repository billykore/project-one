# API Contract: Feature Flags

**Date**: 2026-09-08 | **Feature**: `005-feature-flags`

All endpoints are JSON. Errors use the project's RFC 9457 Problem Details format (see `specs/003-rfc9457-error-format/`). Admin endpoints require a valid session (`access_token` cookie or `Authorization: Bearer`) **and** operator allowlist membership; non-operators receive `403` with no flag detail.

## Base Path

```
/admin/feature-flags      # operator administration
/feature-flags/evaluate   # guarded-feature evaluation
```

---

## Evaluation (used by BFF and protected handlers)

### `GET /feature-flags/evaluate?keys=a,b,c`

Returns a boolean decision for each requested key in the current environment. Optional authentication: with a session, per-user rollout and overrides apply; without one, the global/safe-default result is returned (FR-011).

**Request**

| Param | Type | Required | Notes |
|-------|------|----------|-------|
| `keys` | string (comma-separated) | yes | Flag keys to evaluate; unknown keys return their query but `enabled=false` with `source="unknown"` |

**Response `200`**

```json
{
  "environment": "production",
  "decisions": [
    { "key": "new-feed-algorithm", "enabled": true,  "source": "gradual" },
    { "key": "beta-editor",        "enabled": false, "source": "safe_default" }
  ]
}
```

**Errors**: `400` (missing/invalid `keys`), `422` (too many keys — capped at 50).

---

## Administration (operator-only)

### `GET /admin/feature-flags`

List all flags with a per-environment summary.

**Response `200`**

```json
{
  "flags": [
    {
      "key": "new-feed-algorithm",
      "name": "New Feed Algorithm",
      "owner": "team-feed",
      "lifecycle": "active",
      "safeDefault": false,
      "createdAt": "2026-09-08T10:00:00Z",
      "settings": [
        { "environment": "production", "mode": "gradual", "rolloutPercentage": 25, "revision": 3 }
      ]
    }
  ]
}
```

### `POST /admin/feature-flags`

Create a flag.

**Request**

```json
{
  "key": "new-feed-algorithm",
  "name": "New Feed Algorithm",
  "purpose": "Gradually roll out the ranking rewrite behind the feed.",
  "owner": "team-feed",
  "safeDefault": false
}
```

**Response `201`**: full flag object (as in GET detail). **Errors**: `409` (duplicate key), `422` (validation).

### `GET /admin/feature-flags/{key}`

Flag detail including settings, overrides, and the 20 most recent audit entries.

### `PUT /admin/feature-flags/{key}`

Update metadata (name, purpose, owner, safeDefault). Key and lifecycle are immutable here.

### `PATCH /admin/feature-flags/{key}/environment/{environment}`

Set availability for one environment. Optimistic concurrency via `revision`.

**Request**

```json
{ "mode": "gradual", "rolloutPercentage": 25, "revision": 3, "reason": "Validate with 25% of users" }
```

**Response `200`**: updated setting (new `revision`). **Errors**: `409` (revision mismatch — retry against current state), `422`.

### `PUT /admin/feature-flags/{key}/overrides`

Replace overrides for one environment.

**Request**

```json
{
  "environment": "production",
  "include": ["alice", "bob"],
  "exclude": ["carol"],
  "reason": "Internal validation and support case"
}
```

**Errors**: `422` (a username in both lists, or unknown override type).

### `POST /admin/feature-flags/{key}/archive`

Archive the flag.

**Request**

```json
{ "reason": "Rolled out fully; legacy path removed" }
```

**Response `200`**: flag now `lifecycle: "archived"`. **Errors**: `409` (already archived), `422`.

### `GET /admin/feature-flags/{key}/audit?cursor=...&limit=20`

Cursor-paginated audit history (newest first).

**Response `200`**

```json
{
  "items": [
    {
      "field": "rollout_percentage",
      "environment": "production",
      "previousValue": "10",
      "newValue": "25",
      "actor": "operator1",
      "reason": "Validate with 25% of users",
      "createdAt": "2026-09-08T11:30:00Z"
    }
  ],
  "nextCursor": null,
  "hasMore": false
}
```

---

## Error Summary

| Code | Meaning |
|------|---------|
| `400` | Malformed request (missing `keys`, bad JSON) |
| `401` | Missing/invalid session |
| `403` | Authenticated but not an allowed operator (admin endpoints) |
| `404` | Unknown flag key |
| `409` | Duplicate key, revision mismatch, or illegal transition |
| `422` | Validation failure (RFC 9457 `validation-error`) |
