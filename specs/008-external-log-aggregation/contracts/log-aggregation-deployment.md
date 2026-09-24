# Log Aggregation Deployment Contract

This feature adds no public application endpoint. Its external contract is the deployment wiring between the backend, Alloy, Loki, and Grafana.

## Service boundary

| Component | Responsibility | Network rule |
|---|---|---|
| Backend | Writes structured JSON to stderr | Does not connect to Loki |
| Alloy | Reads only the backend container, exports safe events, and owns best-effort delivery | Requires Docker log-source access and access to Loki |
| Loki | Stores the private backend log stream | Its ingestion port is not published to the host in local Compose |
| Grafana | Searches the Loki stream through a provisioned datasource | Uses Loki's internal base URL, not an ingestion endpoint |

## Configuration boundary

| Input | Local Compose behavior | Protected external deployment behavior |
|---|---|---|
| `APP_ENV` | Supplies the `environment` stream label | Supplies an approved deployment environment value |
| Loki base URL | Uses the internal Loki service | Points to an authenticated TLS proxy or managed Loki endpoint |
| Tenant identity | Omitted unless the destination requires it | Supplied only through deployment configuration |
| Loki credential | Not needed for the private local service | Mounted read-only for Alloy; never committed, displayed, or logged |
| Aggregation enablement | Alloy running means enabled; Alloy absent or stopped means disabled | Controlled by deployment lifecycle; backend logging is unaffected |

The backend has no Loki URL or credential setting. This keeps delivery failures out of application request handling. The existing `deployments/observability/secrets/` location is the local convention for ignored secret files.

## Export contract

- Alloy accepts only JSON stderr records from the backend container.
- It exports the fields and labels defined in [data-model.md](../data-model.md); all other input attributes are discarded before delivery.
- It must generate a safe HTTP completion event from the resolved route template, HTTP method, response status, duration, and request ID. It must not export a raw URI, query string, request body, authorization value, cookie, IP address, user agent, account identifier, raw error, or stack trace.
- Delivery and configuration failures appear only in Alloy's local diagnostics and must not contain credentials or user-provided values; backend stderr remains available independently.
- The initial stream is best-effort. Loki unavailability may lose records but must not delay normal backend workflows.

## Operator query contract

Operators select the deployment stream with the three static labels, then parse safe JSON fields in Grafana Explore. For example, incident correlation uses the `request_id` JSON field after parsing rather than a label matcher. The datasource name is `Loki`; the existing Prometheus datasource remains the default.
