# Tasks: System Health Metrics

**Input**: Design documents from `/specs/007-system-health-metrics/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [observability HTTP contract](./contracts/observability-http.md), [quickstart.md](./quickstart.md)

**Tests**: Required by the project constitution and the feature's independently testable user scenarios. Write each named test first and confirm it fails before its implementation task.

**Organization**: Tasks are grouped by user story so readiness, authenticated operational measurements, and privacy safeguards can be verified incrementally.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel with other marked tasks in its phase because it changes different files and has no incomplete dependency.
- **[Story]**: Maps a task to the corresponding user story in [spec.md](./spec.md).

## Phase 1: Setup

**Purpose**: Add the only new Go dependency required for instrumentation.

- [X] T001 Add the official Prometheus Go client dependency to `go.mod` and `go.sum`.

---

## Phase 2: Foundational Health Architecture

**Purpose**: Establish the Clean Architecture contracts required before HTTP health or metrics work begins.

- [X] T002 Create pure `HealthAssessment` and `ComponentHealth` entities with explicit readiness states in `internal/core/domain/health.go`.
- [X] T003 Add dependency-checker, health-observer, HTTP-metrics, and broker-health contracts in `internal/core/ports/health.go` and `internal/core/ports/pubsub.go`.
- [X] T004 Regenerate GoMock implementations for the new ports in `internal/core/ports/mocks/` by running `make mocks`.

**Checkpoint**: Health assessment can be orchestrated through ports without API handlers or use cases importing adapters.

---

## Phase 3: User Story 1 - Determine Service Readiness (Priority: P1) 🎯 MVP

**Goal**: Give the deployment monitor separate liveness and readiness signals, with a non-sensitive status report for PostgreSQL and the required notification subscriber.

**Independent Test**: Inject successful, failed, and slow database/broker checkers into the health use case. Verify `/healthz` stays live and `/status` returns the contracted component report with `503` whenever any required check is not ready.

### Tests for User Story 1

- [X] T005 [P] [US1] Create failing aggregate-state, timeout, observer, and component-timestamp tests for `internal/core/usecase/health_usecase_test.go` using generated health-port mocks.
- [X] T006 [P] [US1] Create failing liveness, ready, failed-component, and timeout HTTP tests in `internal/api/handler/health_handler_test.go`.
- [X] T007 [P] [US1] Create failing PostgreSQL ping and notification-subscriber health-adapter tests in `internal/adapters/health/health_checker_test.go`.

### Implementation for User Story 1

- [X] T008 [US1] Implement dependency aggregation, bounded check contexts, readiness selection, and health-observer notification in `internal/core/usecase/health_usecase.go`.
- [X] T009 [US1] Implement PostgreSQL and `HealthReporter`-backed notification checkers in `internal/adapters/health/health_checker.go`.
- [X] T010 [US1] Implement HTTP-only liveness/readiness request mapping and non-sensitive response serialization in `internal/api/handler/health_handler.go`.
- [X] T011 [US1] Compose the health checkers, Prometheus-independent health use case, and `/healthz` plus readiness-preserving `/status` routes in `cmd/main.go`.
- [X] T012 [US1] Add Swagger annotations for liveness and readiness responses in `internal/api/handler/health_handler.go`.
- [X] T013 [US1] Run the focused readiness suites in `internal/core/usecase/health_usecase_test.go`, `internal/adapters/health/health_checker_test.go`, and `internal/api/handler/health_handler_test.go`.

**Checkpoint**: `/healthz` proves process liveness; `/status` correctly accepts or rejects normal traffic and identifies only the affected non-sensitive component.

---

## Phase 4: User Story 2 - Observe Operating Trends (Priority: P1)

**Goal**: Let Prometheus collect authenticated operational metrics and let the provisioned Grafana dashboard show request trends, readiness, dependencies, and process start time.

**Independent Test**: Generate successful, 4xx, and 5xx requests for known route templates; verify Prometheus receives the expected bounded series with valid monitoring credentials and Grafana loads the required panels from a healthy target.

### Tests for User Story 2

- [X] T014 [P] [US2] Create failing monitoring username, password-file, missing-secret, and environment-binding tests in `internal/config/config_test.go`.
- [X] T015 [P] [US2] Create failing Prometheus counter, histogram, readiness-gauge, dependency-gauge, and process-collector tests in `internal/adapters/metrics/prometheus_test.go`.
- [X] T016 [P] [US2] Create failing route-template, status-class, unmatched-route, and self-scrape exclusion tests in `internal/api/middleware/metrics_test.go`.

### Implementation for User Story 2

- [X] T017 [US2] Add validated monitoring credential-file configuration and environment bindings in `internal/config/config.go` and `configs/config.yaml.example`.
- [X] T018 [US2] Implement the Prometheus registry, `projectone_http_requests_total`, `projectone_http_request_duration_seconds`, `projectone_dependency_up`, `projectone_ready`, standard collectors, metrics handler, and health-observer adapter in `internal/adapters/metrics/prometheus.go`.
- [X] T019 [US2] Implement HTTP observation through the metrics port using only method, resolved route template (or `unmatched`), and status class in `internal/api/middleware/metrics.go`.
- [X] T020 [US2] Register the authenticated Prometheus handler, bounded-label middleware, and health observer in `cmd/main.go`, preserving error-handler status codes and excluding `/metrics` from self-observation.
- [X] T021 [P] [US2] Create the `project-one-api` scrape target with `/metrics` basic authentication through `password_file` in `deployments/prometheus.yml`.
- [X] T022 [P] [US2] Provision the Prometheus datasource and dashboard provider in `deployments/grafana/provisioning/datasources/project-one-prometheus.yml` and `deployments/grafana/provisioning/dashboards/project-one-health.yml`.
- [X] T023 [US2] Build readiness, dependency, request-rate, 4xx/5xx-rate, p95-duration, and process-start panels in `deployments/grafana/dashboards/project-one-health.json` using `specs/007-system-health-metrics/data-model.md`.
- [X] T024 [US2] Add the ignored local monitoring-secret path in `.gitignore` and wire the backend, Prometheus, Grafana, secrets, volumes, provisioning mounts, service discovery, and Grafana port `3001` in `deployments/compose.yml`.
- [X] T025 [US2] Run the focused configuration, Prometheus-adapter, and middleware suites in `internal/config/config_test.go`, `internal/adapters/metrics/prometheus_test.go`, and `internal/api/middleware/metrics_test.go`.

**Checkpoint**: Prometheus authenticates successfully, scrapes the API, and a fresh Compose deployment opens a populated Project One Health dashboard without manual datasource or panel creation.

---

## Phase 5: User Story 3 - Diagnose Health Without Exposing Users (Priority: P2)

**Goal**: Ensure health and operational information is useful to on-call responders without leaking account data, credentials, request content, raw URLs, or raw dependency errors.

**Independent Test**: Send requests containing representative tokens, account values, and URLs with identifiers; cause a dependency failure; then inspect the health body and Prometheus exposition to confirm the failed component and time are present while sensitive values are absent and unauthenticated metrics are rejected.

### Tests for User Story 3

- [X] T026 [P] [US3] Add privacy regression cases proving raw URLs, query strings, request IDs, credentials, tokens, and user content never become metric labels in `internal/api/middleware/metrics_test.go`.
- [X] T027 [P] [US3] Add privacy regression cases proving readiness responses omit raw database and broker errors while retaining component state and timestamps in `internal/api/handler/health_handler_test.go`.
- [X] T028 [P] [US3] Add missing-credential, invalid-credential, and valid-monitoring-credential route tests for `/metrics` in `cmd/main_test.go`.

### Implementation for User Story 3

- [X] T029 [US3] Apply any failing redaction and route-normalization fixes in `internal/api/handler/health_handler.go` and `internal/api/middleware/metrics.go`.
- [X] T030 [US3] Apply any failing credential-validation or secret-loading fixes in `internal/config/config.go`, `internal/adapters/metrics/prometheus.go`, and `cmd/main.go`.
- [X] T031 [US3] Document monitoring-secret setup, authenticated metrics access, and the continued non-sensitive `/status` probe in `deployments/README.md`.

**Checkpoint**: An incident responder can identify readiness failures safely; unauthenticated clients receive no metrics; and the observability surface contains no user-controlled or secret values.

---

## Phase 6: Polish & Cross-Cutting Validation

**Purpose**: Document the completed stack and verify the application, generated API docs, and Compose workflow together.

- [X] T032 [P] Document Prometheus, Grafana, liveness, readiness, dashboard URL, and monitoring-secret prerequisites in `README.md`.
- [X] T033 Regenerate and review generated API documentation in `api/swagger/docs.go`, `api/swagger/swagger.json`, and `api/swagger/swagger.yaml` after final route implementation.
- [X] T034 Run `gofmt`, `make mocks`, `make test`, and `make check` against `cmd/main.go`, `internal/core/usecase/health_usecase.go`, `internal/adapters/health/health_checker.go`, `internal/adapters/metrics/prometheus.go`, and `internal/api/middleware/metrics.go`.
- [X] T035 [P] Create `scripts/verify-observability-latency.py` to issue 100 authenticated `/metrics`, `/healthz`, and `/status` requests and fail unless at least 99 complete within five seconds.
- [ ] T036 Run every manual end-to-end scenario in `specs/007-system-health-metrics/quickstart.md`, including the latency script, broker outage/recovery, unauthenticated metrics rejection, Prometheus target health, Grafana panels, and privacy inspection.

> **T036 status**: blocked, not executed. The Docker daemon is not running in this environment
> (`docker info` cannot reach the engine socket), so no Compose stack could be started. Everything
> that can be validated without a live stack has been: `make test` and `make check` pass, the
> credential paths in `deployments/compose.yml` were checked against `deployments/prometheus.yml`,
> `docker compose config` resolves the stack, and every dashboard PromQL metric name resolves to a
> metric the adapter exposes. Run T036 after starting Docker:
>
> ```bash
> mkdir -p deployments/observability/secrets
> printf '%s' '<monitoring-password>' > deployments/observability/secrets/metrics-password
> chmod 644 deployments/observability/secrets/metrics-password
> export GRAFANA_ADMIN_PASSWORD='<local-grafana-password>'
> make compose-up
> curl -i http://localhost:8080/healthz
> curl -i http://localhost:8080/status
> curl -u projectone-metrics:'<monitoring-password>' -s http://localhost:8080/metrics | head
> curl -o /dev/null -s -w '%{http_code}\n' http://localhost:8080/metrics   # 401
> MONITORING_PASSWORD='<monitoring-password>' scripts/verify-observability-latency.py
> ```
>
> Then check <http://localhost:9090/targets> for an `UP` `project-one-api` target and
> <http://localhost:3001> for the provisioned Project One Health dashboard.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: T001 has no dependencies.
- **Foundational (Phase 2)**: T002–T004 depend on T001 and establish the domain, ports, and mocks required by the health use case.
- **User Story 1 (Phase 3)**: Depends on T002–T004. It is the MVP and can be delivered before Prometheus/Grafana.
- **User Story 2 (Phase 4)**: Depends on T001 and the health use case from US1 for readiness gauges. T014–T016 and T021–T022 may proceed in parallel after their prerequisites.
- **User Story 3 (Phase 5)**: Depends on the public surfaces from US1 and US2 so privacy and credential tests inspect completed responses and metrics.
- **Polish (Phase 6)**: Depends on all desired user-story phases being complete.

### User Story Dependencies

```text
Setup ──> Foundational ──> US1 (readiness MVP) ──> US2 (authenticated metrics + dashboard) ──> US3 (privacy hardening) ──> Polish
```

- **US1** has no user-story dependency and is independently deployable after Foundational.
- **US2** can prepare its tests and deployment assets after Setup; it needs US1 complete before readiness and dependency gauges are meaningful.
- **US3** verifies the completed US1 and US2 output surfaces, then adds access-control and secret-safety documentation.

### Parallel Opportunities

- In **US1**, T005–T007 can be written in parallel because they test different layers.
- In **US2**, T014–T016 and T021–T022 can proceed in parallel; T018/T019/T020 remain sequential where they share runtime wiring.
- In **US3**, T026–T028 can proceed in parallel; T031 is independent of the test files.
- T032 can be drafted in parallel with late validation because it changes only `README.md`.

## Parallel Example: User Story 2

```text
T014: internal/config/config_test.go
T015: internal/adapters/metrics/prometheus_test.go
T016: internal/api/middleware/metrics_test.go
T021: deployments/prometheus.yml
T022: deployments/grafana/provisioning/datasources/project-one-prometheus.yml
```

After those complete, implement T017–T020 in order, then finish the dashboard and Compose integration.

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete T001–T004.
2. Complete T005–T013.
3. Validate `/healthz` and `/status` against the independent test criterion.
4. Deliver the Clean-Architecture readiness MVP before adding metrics infrastructure.

### Incremental Delivery

1. Deliver US1 for safe deployment routing and dependency diagnosis.
2. Deliver US2 for authenticated Prometheus collection and the Grafana dashboard.
3. Deliver US3 for privacy and credential regressions.
4. Complete T032–T036 before merging.
