# Implementation Plan: External Log Aggregation

**Branch**: `008-external-log-aggregation` | **Date**: 2026-09-23 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/008-external-log-aggregation/spec.md`; use Loki and Grafana as the aggregation system.

## Summary

Extend the existing Compose observability stack with Loki as the private log store, Grafana Alloy as the out-of-process collector, and a provisioned Loki datasource in the existing Grafana instance. The Go application continues to emit JSON records to stderr; Alloy reads only the backend container, transforms each record into an allowlisted safe event, and delivers it asynchronously to Loki. A small, safe HTTP completion event supplies request correlation without forwarding raw URLs, query strings, addresses, or user agents. Aggregation enablement is deployment-owned: Alloy's running state is the enabled/disabled signal, Alloy writes delivery failures to its own local diagnostics, and backend stderr remains independent. Grafana Explore is the operator search interface; no application UI or custom log dashboard is added.

## Technical Context

**Language/Version**: Go 1.26.2; Docker Compose YAML; Grafana Alloy configuration

**Primary Dependencies**: Existing Go `log/slog`, Echo 4.15.1, Viper, Docker Compose, Grafana 12.2.0; Loki and Grafana Alloy container images. No new Go dependency.

**Storage**: Existing local stderr logs; named Loki and Alloy state volumes; existing Grafana configuration volume. No application database or migration changes.

**Testing**: Go standard test runner with Testify/GoMock where applicable; pre-change and final backend coverage comparison; frontend Vitest; a standard-library read/write p95 verifier; `docker compose ... config`; Compose-based 100-event and 20-correlation-context validation; a prohibited-value sentinel matrix across Loki and Alloy diagnostics; and a simulated Loki outage.

**Target Platform**: Linux containers run through Docker Compose; existing Go API on port 8080.

**Project Type**: Full-stack web application; this feature changes backend logging and deployment observability only.

**Performance Goals**: Existing API p95 limits remain below 200 ms for reads and 500 ms for writes. During normal operation, all 100 representative events are searchable in Loki within 30 seconds; asynchronous collection never waits on Loki in a request path.

**Constraints**: Alloy reads only the `backend` container and sends one bounded-label stream to Loki. No account identifiers, credentials, tokens, cookies, request bodies, message content, raw URLs/query strings, IP addresses, user agents, raw errors, or stack traces may be exported or echoed in Alloy delivery diagnostics. `request_id` and instance identity are structured event fields, never Loki labels. Loki has no published ingestion port in the local stack. When Alloy is absent or stopped, Compose status provides the disabled-state signal and backend stderr continues unchanged.

**Scale/Scope**: One backend service and one collector per Compose deployment; static stream labels `app=project-one`, `environment`, and `source=backend`; one Loki datasource in the existing Grafana. No durable replay/WAL, alerting, tracing, multi-destination routing, or custom dashboard.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Pre-design gate — PASS**

- **Clean Architecture**: No domain rule or use case gains a Loki dependency. Logging remains in the existing adapter; the API layer owns the HTTP completion event; all Loki, Alloy, and Grafana integration remains deployment infrastructure.
- **Testing**: Capture pre-change backend coverage, add focused tests for safe completion-event fields, and require the final coverage total not to decrease. Validate Alloy output with known safe and prohibited values; run `make test`, `make check`, `npm test -- --run`, and the Compose quickstart.
- **UX**: The Next.js UI is unchanged. Grafana Explore, already part of the selected operational system, is the search interface; no user-facing log screen is introduced.
- **Performance**: The collector is outside the Go request path. Its batching, retry, and backpressure cannot delay backend requests. The completion logger uses the resolved route template rather than raw request data. A repeatable acceptance verifier enforces read p95 below 200 ms and write p95 below 500 ms under normal local load.
- **Security**: An allowlist is applied before records leave the backend container. Loki stays internal; production deployments must put an authenticated TLS proxy in front of it and supply collector credentials through secrets, never source-controlled values.

**Post-design gate — PASS**

The design reuses the application's JSON stderr logger and its existing Compose/Grafana provisioning patterns. No new Go client, core port, schema, application endpoint, or frontend package is needed. The only Go change emits a safe, bounded HTTP completion record; Alloy independently filters every exported record, so existing local diagnostic logging does not leak to Loki.

## Project Structure

### Documentation (this feature)

```text
specs/008-external-log-aggregation/
├── plan.md
├── research.md
├── data-model.md
├── coverage-baseline.txt             # Pre-change backend coverage total
├── quickstart.md
├── contracts/
│   └── log-aggregation-deployment.md
└── tasks.md                         # Created by /speckit-tasks
```

### Source Code (repository root)

```text
cmd/
└── main.go                           # Replace default request logging with safe completion logging

internal/
└── api/middleware/
    └── request_logging.go            # Safe resolved-route completion event and tests

deployments/
├── compose.yml                       # Loki and Alloy services, internal wiring, named volumes
├── loki.yml                          # Single-process local Loki configuration
├── alloy/config.alloy                # Backend-only Docker discovery, sanitization, and Loki delivery
└── grafana/provisioning/datasources/
    └── project-one-loki.yml          # Provisioned Loki datasource

README.md                             # Local observability entry points
deployments/README.md                 # Loki, Grafana Explore, secrets, and outage guidance

scripts/
└── verify-request-latency.py          # Read/write p95 acceptance check
```

**Structure Decision**: Extend the existing backend logger and deployment observability stack in place. Grafana Alloy and Loki are Compose services, not application packages; the frontend, core layers, database, HTTP contract, and Makefile remain unchanged.
