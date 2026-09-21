# Implementation Plan: System Health Metrics

**Branch**: `feat/app-metrics` | **Date**: 2026-09-21 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/007-system-health-metrics/spec.md`

## Summary

Expose application liveness, dependency readiness, and privacy-safe HTTP operational metrics. A health use case aggregates database and broker health ports; the HTTP handler only maps that result to liveness/readiness responses. A Prometheus adapter records bounded-label metrics and serves authenticated metrics; Prometheus scrapes them with a Compose secret; Grafana loads one versioned health dashboard. The existing `/status` probe becomes the readiness report, while a separate liveness route distinguishes a running process from one ready to serve traffic.

## Technical Context

**Language/Version**: Go 1.26.2; YAML and JSON for deployment/provisioning assets

**Primary Dependencies**: Echo 4.15.1; official `github.com/prometheus/client_golang`; Prometheus and Grafana container images

**Storage**: Existing PostgreSQL application store; Prometheus local time-series volume and Grafana local configuration volume; a Compose-mounted monitoring secret; no application schema changes

**Testing**: Go standard test runner with Testify/GoMock where applicable; Compose-based manual integration validation

**Target Platform**: Linux containers run through Docker Compose; existing Go API on port 8080

**Project Type**: Full-stack web application; this feature changes the backend and deployment observability stack only

**Performance Goals**: 99% of health and metric responses within five seconds; existing API p95 limits remain below 200 ms for reads and 500 ms for writes under normal load

**Constraints**: Bounded metric labels only; no identifiers, tokens, payloads, request IDs, or error text in metric labels; readiness checks use a bounded context; `/metrics` requires a dedicated monitoring credential read from a deployment secret

**Scale/Scope**: One API service, PostgreSQL, and the configured notification broker; one Prometheus scrape job and one provisioned Grafana dashboard; no alerts, tracing, retention policy, or automated remediation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Pre-design gate — PASS**

- **Clean Architecture**: Health assessment is an application use case over health-check and metrics-observer ports; database, broker, and Prometheus implementations live in adapters. API handlers only map requests and results. No use case imports an adapter.
- **Testing**: Add focused use-case, adapter, middleware, and health-handler tests; regenerate GoMock mocks after new ports. Run `make test`; the Compose quickstart exercises the complete stack.
- **UX**: No end-user application UI changes. The Grafana dashboard is an internal operational interface and will be provisioned rather than coupled to the Next.js application.
- **Performance**: Request observations are constant-time counter/histogram updates with bounded labels. Database readiness has a short context deadline; metric scraping is excluded from its own HTTP instrumentation.
- **Security**: `/status` remains non-sensitive for the deployment probe. Detailed metrics contain only bounded operational labels and require dedicated monitoring credentials stored outside source control.

**Post-design gate — PASS**

The design keeps Prometheus-specific code in an adapter behind ports, gives the health use case direct mock seams for each dependency state, preserves the existing deployment probe, and introduces no data migration, frontend bundle, or unbounded query. No constitution exception is required.

## Project Structure

### Documentation (this feature)

```text
specs/007-system-health-metrics/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
```text
cmd/
└── main.go                                  # Application composition and route wiring

internal/
├── api/
│   ├── handler/
│   │   └── health_handler.go                # HTTP liveness/readiness mapping and tests
│   └── middleware/
│       └── metrics.go                       # HTTP observation through a metrics port and tests
├── adapters/
│   ├── health/                              # PostgreSQL and notification health-check adapters
│   ├── metrics/                             # Prometheus recorder, registry, and metrics handler
│   └── pubsub/                              # Existing broker healthy-state implementations and tests
├── config/
│   └── config.go                            # Monitoring credential-file configuration
└── core/
    ├── domain/
    │   └── health.go                        # Pure health-assessment entities
    ├── ports/
    │   ├── health.go                        # Dependency and metrics-observer contracts
    │   └── pubsub.go                        # Existing broker healthy-state capability
    └── usecase/
        └── health_usecase.go                # Health assessment orchestration and tests

deployments/
├── compose.yml                              # Prometheus and Grafana services
├── prometheus.yml                           # API scrape job
├── observability/secrets/                   # Ignored local monitoring secret file
└── grafana/
    ├── provisioning/
    │   ├── dashboards/                      # Dashboard-provider configuration
    │   └── datasources/                     # Prometheus datasource configuration
    └── dashboards/
        └── project-one-health.json          # Versioned health dashboard

scripts/
└── verify-observability-latency.py           # Reproducible health and metrics latency acceptance check
```

**Structure Decision**: Extend the existing Go API and Compose deployment in place. Prometheus and Grafana are deployment services, not new application projects; the Next.js frontend remains unchanged.
