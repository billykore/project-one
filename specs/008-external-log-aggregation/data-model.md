# Data Model: External Log Aggregation

## Exported log event

The collector emits only this safe projection of a backend JSON record. Missing optional fields are omitted; no raw record is forwarded.

| Field | Type | Description | Validation |
|---|---|---|---|
| `timestamp` | RFC 3339 timestamp | Time at which the backend recorded the event | Required; derived from the backend JSON record |
| `level` | enum | Event severity | Required; one of the standard application severities |
| `message` | string | Stable, human-readable event name | Required; must not contain user-provided content |
| `request_id` | string | Correlates a completed HTTP request with its client-visible problem detail when applicable | Optional; JSON field only, never a label |
| `method` | enum | HTTP method for a completion event | Optional; only a standard HTTP method |
| `route` | string | Resolved route template for a completion event | Optional; no raw URL, query string, or route parameter value |
| `status` | integer | Completed HTTP response status | Optional; valid HTTP status |
| `duration_ms` | number | Completed HTTP request duration | Optional; non-negative |
| `error_code` | string | Stable application error category | Optional; must not contain raw error text |
| `failure_category` | string | Stable operational failure category | Optional; must not contain user-provided content |
| `instance_id` | string | Identity of the currently running backend container | Required when supplied by the collector; JSON field only |

The collector may use a secret-filtering stage as defense in depth. It must not treat that stage as a substitute for the allowlist.

## Loki stream labels

| Label | Value | Rule |
|---|---|---|
| `app` | `project-one` | Fixed for this application |
| `environment` | Deployment environment | One of the deployment's approved environment values |
| `source` | `backend` | Fixed; the collector excludes every other Compose service |

No other labels are permitted in the initial release. In particular, request IDs, instance IDs, users, account IDs, paths, query values, client addresses, errors, and timestamps are not labels.

## Delivery configuration

| Field | Type | Description | Validation |
|---|---|---|---|
| `loki_url` | URL | Base URL used by Alloy to deliver events | Required for an external destination; local Compose defaults to the private Loki service |
| `environment` | enum | Deployment identity injected into stream labels | Required; must match the deployment environment |
| `tenant_id` | optional string | Deployment-provided Loki tenant identity when required by the protected destination | Never logged or labelized |
| `credential_file` | optional secret-file reference | Read-only credential supplied to Alloy for a protected destination | File contents are never read into source-controlled configuration or logs |

## Lifecycle

```text
backend JSON stderr record
        │
        ▼
Alloy discovers backend container ──> parse record ──> allowlisted safe event
        │                                                    │
        │ Loki unavailable                                    ▼
        └──> backend still serves requests              Loki stream
                                                          │
                                                          ▼
                                                 Grafana Explore query
```

During a Loki or Alloy outage, local backend stderr logging continues and the backend keeps serving users. The collector resumes delivery for later records after recovery; the initial release does not guarantee recovery of records that could not be delivered.
