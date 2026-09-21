# Data Model: System Health Metrics

## Health report

| Field | Type | Description | Validation |
|---|---|---|---|
| `status` | enum | Overall readiness: `ready` or `not_ready` | `ready` only when every required component is `up` |
| `checked_at` | timestamp | UTC time the assessment completed | Required; RFC 3339 representation |
| `components` | list of Component Health | Assessed dependencies | Contains `database` and the enabled notification broker |

## Component health

| Field | Type | Description | Validation |
|---|---|---|---|
| `name` | enum | Stable component name | `database`, `rabbitmq`, `kafka`, or `inmemory` as configured |
| `status` | enum | Current check result | `up`, `down`, or `unknown` |
| `checked_at` | timestamp | UTC time this component was assessed | Required |

The public report deliberately omits raw connection errors, addresses, credentials, queue names, and broker responses. The server logs may retain diagnostic errors through the established structured logger.

## Health state transitions

```text
process starts ──> liveness up
                     │
                     ├── database ping succeeds AND subscriber healthy ──> readiness ready
                     │
                     └── a required check fails, times out, or is unknown ──> readiness not_ready
                                                                          │
                                                                          └── later check succeeds ──> readiness ready
```

Liveness is independent of readiness: a running process can be live while it is not ready to receive normal traffic.

## Prometheus metric families

| Metric | Type | Labels | Meaning |
|---|---|---|---|
| `projectone_http_requests_total` | Counter | `method`, `route`, `status_class` | Completed application requests, excluding the metrics scrape itself |
| `projectone_http_request_duration_seconds` | Histogram | `method`, `route` | Completed application request duration in seconds |
| `projectone_dependency_up` | Gauge | `dependency` | `1` when the latest readiness check found the component up, otherwise `0` |
| `projectone_ready` | Gauge | None | `1` when the latest readiness assessment is ready, otherwise `0` |
| `process_start_time_seconds` | Gauge | Default collector labels only | Start time of the running Go process |
| `go_*`, `process_*` | Standard collectors | Collector-defined bounded labels | Go runtime and process observations supplied by the official client |

### Label rules

- `route` is Echo's resolved route template, never the request URL. Unmatched requests use a single `unmatched` value.
- `status_class` is one of `1xx`, `2xx`, `3xx`, `4xx`, or `5xx`.
- `dependency` is fixed by the active deployment. It never includes hosts, ports, exchanges, queues, or user data.
- No metric label or metric name may contain account identifiers, request IDs, credentials, tokens, request or message contents, query strings, raw errors, or timestamps.

### Lifecycle and aggregation

- Counters and histograms begin when an application instance starts and reset on restart; `process_start_time_seconds` identifies the reset.
- Prometheus performs retention and cross-instance aggregation. Grafana queries Prometheus rather than receiving measurements from the application directly.
- Health gauges are updated by each readiness request. A failed check updates its affected dependency and overall readiness to `0` while returning the report.

## Monitoring credential

| Field | Type | Description | Validation |
|---|---|---|---|
| `username` | string | Fixed machine identity used only to scrape `/metrics` | Required; not an application account |
| `password_file` | secret file path | Mounted file containing the matching secret | Required; secret content is never logged, returned, or committed |

The application reads the credential from deployment configuration and the Prometheus scrape configuration reads the same mounted secret file. Neither component places the secret in a dashboard, metric label, response body, or source-controlled configuration.
