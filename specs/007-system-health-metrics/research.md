# Research: System Health Metrics

## 1. Prometheus instrumentation

**Decision**: Use the official Prometheus Go client with a single application registry, standard Go/process collectors, a request counter, and a request-duration histogram. Mount its HTTP handler through Echo at `/metrics`.

**Rationale**: The official Go guidance uses `promhttp` to expose metrics and supports a dedicated registry with Go and process collectors. Those collectors already export runtime/process health, including process start time, so the feature does not need to recreate them. One Echo middleware can measure every normal HTTP route while leaving domain and use-case code untouched.

**Alternatives considered**:

- Hand-written metric text: rejected because it duplicates the maintained Prometheus exposition client.
- A third-party Echo metrics wrapper: rejected because a small local middleware gives the required bounded route-template label and adds no wrapper dependency.
- Separate registries per feature: rejected because one registry is simpler to expose and test; isolated registries are only needed if later tests or plugin isolation demand them.

**Sources**: [Instrumenting a Go application](https://prometheus.io/docs/guides/go-application/), [Writing client libraries](https://prometheus.io/docs/instrumenting/writing_clientlibs/).

## 2. Metric names, types, and labels

**Decision**: Use the `projectone_` namespace and only bounded labels: HTTP method, Echo route template, and response status class; dependency name is limited to `database` and the configured broker. Use counters for requests, histograms (seconds) for request duration, and gauges for readiness and dependency state.

**Rationale**: Prometheus recommends an application prefix, one base unit per metric, counters for cumulative events, and histograms for duration. A raw path, account, request ID, token, or error string creates unbounded cardinality and violates the privacy requirement. Route templates retain useful operation grouping without recording user input.

**Alternatives considered**:

- Raw URL or account labels: rejected because values are unbounded and may expose data.
- One metric family per response code or route: rejected because labels express the bounded dimension without dynamic metric names.
- Summaries: rejected because histograms support aggregating percentiles across instances in Prometheus.

**Sources**: [Metric and label naming](https://prometheus.io/docs/practices/naming/), [Instrumentation practices](https://prometheus.io/docs/practices/instrumentation/), [Writing client libraries](https://prometheus.io/docs/instrumenting/writing_clientlibs/).

## 3. Liveness and readiness checks

**Decision**: Keep `/status` as the deployment readiness route and add `/healthz` for process liveness. A health use case performs a short, fresh database check through a port and checks the notification subscriber through a port-backed adapter. The HTTP handler maps the non-sensitive component-level result to HTTP 503 whenever a required check is not ready.

**Rationale**: The existing Docker probe already consumes `/status`; retaining that path minimizes deployment churn while making its status meaningful. The health use case preserves the project's requirement that handlers delegate application logic and that external integrations are ports. The RabbitMQ publisher connects lazily and remains unhealthy until it publishes its first event, so it cannot establish startup readiness. The subscriber becomes healthy after it establishes consumption and becomes unhealthy on reconnect/close, directly representing notification delivery availability.

**Alternatives considered**:

- Keep static `/status`: rejected because it cannot identify dependency failures.
- Use the publisher as broker readiness: rejected because its lazy connection reports a false negative before the first publish.
- Probe a broker by publishing a synthetic event: rejected because it mutates delivery state and risks user-visible side effects.
- Add a new external health service: rejected because the application already has all required dependency state.

## 4. Prometheus scrape boundary

**Decision**: Prometheus scrapes `backend:8080/metrics` over the internal Compose network with dedicated HTTP basic-auth monitoring credentials mounted as a deployment secret. The application does not couple metrics access to user sessions or JWTs.

**Rationale**: Prometheus uses a pull scrape model and supports an explicit scrape target, path, and basic-auth credential file. Keeping the endpoint on the existing API listener avoids another listener, while machine credentials enforce access even when the API listener is reachable outside the Compose network. The secret is independent from user authentication.

**Alternatives considered**:

- Public metrics endpoint: rejected because it conflicts with the trusted-operations requirement.
- Require an application user session or JWT: rejected because Prometheus is an automated client and this would create unnecessary identity coupling.
- Network-only restriction: rejected because the current API listener can be exposed by a deployment port mapping and documentation alone cannot enforce FR-013.
- Add a second metrics listener: deferred unless a deployment requires stronger network separation than dedicated credentials provide.

**Sources**: [Prometheus configuration](https://prometheus.io/docs/prometheus/latest/configuration/configuration/), [Instrumenting a Go application](https://prometheus.io/docs/guides/go-application/).

## 5. Grafana provisioning

**Decision**: Add Grafana as a Compose service with a provisioned Prometheus datasource and one versioned dashboard. The dashboard covers readiness, dependency state, request rate, client/server error rate, request-duration percentiles, and instance start time. No alert rules or notification policies are provisioned.

**Rationale**: Grafana includes the Prometheus datasource and supports datasource and dashboard provisioning from files. Versioned provisioning makes a new local environment immediately useful and avoids dashboard drift from untracked UI edits. This follows the user's explicit Grafana instruction while keeping alerts outside the agreed scope.

**Alternatives considered**:

- Use Grafana Explore only: rejected because the user explicitly selected Grafana and the specification now requires a ready dashboard.
- Build a dashboard in the Next.js application: rejected because Grafana already provides the operational visualization layer.
- Provision alerts: rejected because alert routing remains out of scope.

**Sources**: [Prometheus data source](https://grafana.com/docs/grafana/latest/datasources/prometheus/), [Configure the Prometheus data source](https://grafana.com/docs/grafana/latest/datasources/prometheus/configure/), [Grafana provisioning](https://grafana.com/docs/grafana/latest/administration/provisioning/).
