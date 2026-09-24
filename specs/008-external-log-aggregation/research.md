# Research: External Log Aggregation

## 1. Collection and delivery architecture

**Decision**: Keep the existing JSON `slog` output to stderr. Use Grafana Alloy to discover and read only the backend Docker container, sanitize its records, and send them to Loki; provision the existing Grafana instance with a Loki datasource.

**Rationale**: The backend already emits structured JSON to stderr, so a collector avoids a new Go dependency and prevents Loki network latency, retries, batching, or backpressure from entering request handling. Alloy supplies Docker discovery and a maintained Loki writer; Grafana already provides the required search experience through Explore.

**Alternatives considered**:

- Direct Loki push from Go: rejected because it would require an application-owned bounded queue, batching, retry classification, credential handling, and shutdown behavior.
- OpenTelemetry logging: rejected because it adds an SDK and collector scope without a tracing requirement.
- A custom Next.js log-search page or Grafana dashboard: rejected because Grafana Explore already searches Loki and the specification excludes a new log-search UI.

**Sources**: [Alloy Docker source](https://grafana.com/docs/alloy/latest/reference/components/loki/loki.source.docker/), [Alloy Loki writer](https://grafana.com/docs/alloy/latest/reference/components/loki/loki.write/), [Loki supported senders](https://grafana.com/docs/loki/latest/send-data/), [Grafana Loki datasource](https://grafana.com/docs/grafana/latest/datasources/loki/configure/).

## 2. Delivery failure behavior

**Decision**: Use Alloy's standard `loki.write` batching and retry behavior without a write-ahead log, durable replay, or experimental queue configuration. Keep Loki and the collector outside the backend request path.

**Rationale**: Best-effort delivery is explicitly in scope; the existing local stderr output remains available if the destination is unavailable. Default writer behavior batches for one second or one MiB, times out a request after ten seconds, retries 429 responses with exponential backoff, and treats other 4xx responses as actionable destination/configuration failures. This meets availability goals without creating a second durability system.

**Alternatives considered**:

- Enable collector write-ahead logging and replay: rejected because the specification excludes durable replay of undelivered events.
- Add a custom retry queue in the Go process: rejected because it would add failure modes to normal user workflows.
- Increase Loki rate limits preemptively: rejected because limits should be tuned only after a validated noisy stream or capacity requirement.

**Sources**: [Alloy writer settings](https://grafana.com/docs/alloy/latest/reference/components/loki/loki.write/), [Loki limits](https://grafana.com/docs/loki/latest/configure/loki-limits/), [Loki rate-limit validation](https://grafana.com/docs/loki/latest/operations/request-validation-rate-limits/).

## 3. Safe event schema and labels

**Decision**: Export an allowlisted JSON event body and use only static, bounded Loki labels: `app=project-one`, `environment`, and `source=backend`. Keep `request_id`, instance identity, severity, routes, and all other per-event values in the JSON event body, not labels. Add a safe HTTP completion event based on method, resolved route template, status, duration, and request ID.

**Rationale**: Existing logs contain usernames, emails, raw errors, stack traces, and request data that cannot be forwarded unchanged. An allowlist at the collector is the shared enforcement point for all existing calls and preserves local diagnostics. Static labels avoid high cardinality; Grafana queries JSON fields with `| json` when correlating an incident. The completion event gives every completed HTTP request a safe correlation record without the default request logger's raw URI, remote address, or user-agent values.

**Alternatives considered**:

- Forward all JSON records and only remove known secret keys: rejected because user identifiers, arbitrary errors, and new fields could still leak.
- Label streams with request ID, account, route parameter, container ID, or severity: rejected because request and identity values are high-cardinality or sensitive; severity is queryable from event JSON.
- Remove sensitive fields from every existing log call: rejected because it is broad, error-prone, and would unnecessarily change local diagnostics; collector-side allowlisting provides one external boundary.

**Sources**: [Loki label best practices](https://grafana.com/docs/loki/latest/get-started/labels/bp-labels/), [Loki structured metadata](https://grafana.com/docs/loki/latest/get-started/labels/structured-metadata/), [Alloy secret filtering](https://grafana.com/docs/alloy/latest/reference/components/loki/loki.secretfilter/).

## 4. Deployment security and access

**Decision**: Do not publish Loki's ingestion port in Compose. Provision Grafana with Loki's internal base URL. For non-local deployments, require a TLS-authenticated proxy in front of Loki and inject collector credentials as a read-only deployment secret.

**Rationale**: Loki does not provide built-in authentication. Keeping the local service on the Compose network is sufficient for local validation, while production requires an explicit protected boundary. Grafana needs the Loki base URL rather than its ingestion path; datasource provisioning makes the integration reproducible.

**Alternatives considered**:

- Publish Loki directly to the host: rejected because it exposes an unauthenticated ingestion/query service.
- Use one shared credential for Grafana and the collector: rejected because ingestion and query access have different responsibilities.
- Treat Docker socket access as a production authorization model: rejected because it is privileged; production collectors must use the platform's restricted log-source permissions.

**Sources**: [Loki authentication](https://grafana.com/docs/loki/latest/operations/authentication/), [Grafana Loki datasource configuration](https://grafana.com/docs/grafana/latest/datasources/loki/configure/), [Alloy access permissions](https://grafana.com/docs/alloy/latest/access_permissions/).
