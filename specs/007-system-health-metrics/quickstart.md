# Quickstart: System Health Metrics

## Prerequisites

- Docker with Compose and the existing RSA key pair under `configs/keys/`.
- A local Grafana administrator password and a monitoring password. Do not commit either value.

## Start the observability stack

From the repository root:

```bash
export GRAFANA_ADMIN_PASSWORD='replace-with-a-local-secret'
mkdir -p deployments/observability/secrets
printf '%s' 'replace-with-a-monitoring-secret' > deployments/observability/secrets/metrics-password
make compose-up
```

The Compose stack starts the API, Prometheus, and Grafana. Prometheus scrapes the API through the Compose service network; Grafana is provisioned with Prometheus as its datasource and the Project One Health dashboard.

## Validate liveness and readiness

```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/status
```

Expected results:

- `/healthz` returns `200` while the process is running.
- `/status` returns `200` with `status: ready`, plus `database` and `rabbitmq` components, after the notification subscriber has connected.

Stop RabbitMQ to validate the not-ready path, then start it again:

```bash
docker compose -f deployments/compose.yml stop rabbitmq
curl -i http://localhost:8080/status
docker compose -f deployments/compose.yml start rabbitmq
```

The report should return `503` and identify RabbitMQ as not up while it is stopped, then return to `200` after the subscriber reconnects. Normal report requests must still complete promptly.

## Validate metrics and dashboard

Generate a few requests, including one client error:

```bash
curl -o /dev/null -s -w '%{http_code}\n' http://localhost:8080/healthz
curl -o /dev/null -s -w '%{http_code}\n' http://localhost:8080/posts/not-a-number
curl -u projectone-metrics:replace-with-a-monitoring-secret -s http://localhost:8080/metrics | rg 'projectone_(http|dependency|ready)|process_start_time_seconds'
curl -o /dev/null -s -w '%{http_code}\n' http://localhost:8080/metrics
```

The unauthenticated metrics request must return `401`. Open Prometheus at `http://localhost:9090/targets` and confirm that the `project-one-api` target is `UP`. Then open Grafana at `http://localhost:3001`, sign in as `admin` with `GRAFANA_ADMIN_PASSWORD`, and open the provisioned **Project One Health** dashboard. Confirm that it displays readiness, dependency state, request rate, 4xx/5xx rate, p95 duration, and process start time.

## Automated checks

```bash
make test
make check
scripts/verify-observability-latency.py
```

The latency script issues 100 authenticated health and metrics requests and fails unless at least 99 complete within five seconds. The implementation also adds focused tests for metric labels and observations, liveness/readiness states, and configuration/route behavior. Review the [HTTP contract](./contracts/observability-http.md) and [data model](./data-model.md) when diagnosing a failed validation.
