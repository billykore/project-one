# Observability HTTP Contract

## GET `/healthz`

Reports whether the Go process can respond. It does not contact external dependencies.

| Condition | Status | Body |
|---|---|---|
| Process is running | `200 OK` | `{"status":"ok","checked_at":"<RFC3339 UTC timestamp>"}` |

This route is non-sensitive and unauthenticated. It is suitable for a liveness probe but not for routing normal traffic.

## GET `/status`

Reports application readiness. This preserves the existing deployment probe path.

| Condition | Status | Body |
|---|---|---|
| All required checks are up | `200 OK` | Readiness Report |
| Any required check is down, unknown, or times out | `503 Service Unavailable` | Readiness Report |

### Readiness Report

```json
{
  "status": "ready",
  "checked_at": "2026-09-21T12:00:00Z",
  "components": [
    {"name": "database", "status": "up", "checked_at": "2026-09-21T12:00:00Z"},
    {"name": "rabbitmq", "status": "up", "checked_at": "2026-09-21T12:00:00Z"}
  ]
}
```

`status` is `not_ready` whenever any required component is not `up`. The body never contains an address, credential, token, account identifier, request content, or raw dependency error.

## GET `/metrics`

Exposes Prometheus text-format metrics. It is intentionally not part of the public application API or Swagger contract.

| Condition | Status | Body |
|---|---|---|
| Request has valid monitoring credentials | `200 OK` | Prometheus text exposition |
| Request has missing or invalid monitoring credentials | `401 Unauthorized` | No metric content |

Prometheus scrapes the route on the private Compose network at `backend:8080/metrics` with the dedicated monitoring credentials mounted from a deployment secret. The route excludes its own scrape from `projectone_http_*` observations. The credential is not an application user account and must not appear in application configuration, source control, logs, or metrics.

## Compatibility

- `/status` remains unauthenticated and retains its role as the Compose backend health check.
- Its response now carries readiness detail and may return `503` during startup or a required dependency outage; callers that previously only accepted a static `200` must use `/healthz` if they need process liveness rather than readiness.
