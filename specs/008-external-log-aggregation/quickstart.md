# Quickstart: External Log Aggregation

## Prerequisites

- Docker with Compose, the existing RSA key pair under `configs/keys/`, the existing monitoring secret, and `GRAFANA_ADMIN_PASSWORD` as described in [deployments/README.md](../../deployments/README.md).
- A running local Compose stack after the implementation adds Loki and Alloy. Do not publish Loki's ingestion port or commit any protected-destination credential.

## Validate deployment wiring

From the repository root, validate the resolved stack and start it:

```bash
docker compose -f deployments/compose.yml config
make compose-up
docker compose -f deployments/compose.yml ps
```

Expected results:

- Backend, Loki, Alloy, Prometheus, and Grafana are running.
- Loki has no host-published ingestion port.
- Alloy is configured to collect only the backend service.

Aggregation enablement is deployment-owned: Alloy running means enabled; Alloy absent or stopped means disabled. In the disabled state, `docker compose -f deployments/compose.yml ps alloy` shows that the collector is not running while backend stderr remains available.

## Generate and find safe events

Generate a normal request and an error request with known request IDs:

```bash
curl -H 'X-Request-ID: log-normal-001' -o /dev/null -s -w '%{http_code}\n' http://localhost:8080/healthz
curl -H 'X-Request-ID: log-error-001' -o /dev/null -s -w '%{http_code}\n' 'http://localhost:8080/posts/not-a-number'
```

Open Grafana at <http://localhost:3001>, sign in, choose **Explore**, and select the provisioned **Loki** datasource. Query the backend stream, parse JSON fields, and locate `log-error-001`:

```logql
{app="project-one", environment="development", source="backend"} | json | request_id="log-error-001"
```

Expected results:

- Both completion events are searchable within 30 seconds.
- The error event includes only safe correlation and response fields.
- The existing Prometheus datasource and health dashboard remain available.

Generate 100 events containing 20 repeated correlation contexts, then start a two-minute operator lookup timer:

```bash
for event_number in $(seq 1 100); do
  request_number=$(printf '%02d' "$((event_number % 20 + 1))")
  curl -H "X-Request-ID: log-context-${request_number}" -o /dev/null -s http://localhost:8080/healthz
done
```

Confirm all 100 events become searchable within 30 seconds. Using event time, severity, and `request_id`, locate each value from `log-context-01` through `log-context-20` within two minutes.

## Validate privacy boundary

Send distinct sentinel values through prohibited locations:

```bash
curl -H 'X-Request-ID: log-private-001' -o /dev/null -s 'http://localhost:8080/posts/not-a-number?private=LOG_PRIVATE_SENTINEL'
curl -H 'Authorization: Bearer TOKEN_SENTINEL' -H 'Cookie: session=COOKIE_SENTINEL' -o /dev/null -s http://localhost:8080/healthz
curl -H 'Content-Type: application/json' -o /dev/null -s -X POST http://localhost:8080/auth/login -d '{"email":"ACCOUNT_SENTINEL@example.invalid","password":"PASSWORD_SENTINEL","message":"PRIVATE_MESSAGE_SENTINEL"}'
```

In Grafana Explore, search the stream for every sentinel above; all searches must return zero results. Inspect `docker compose -f deployments/compose.yml logs alloy` and confirm it contains none of those values, raw error content, or stack traces.

## Validate destination configuration

Validate each deployment state independently:

```bash
# Default: private local Loki.
docker compose -f deployments/compose.yml up -d loki alloy
curl -H 'X-Request-ID: log-config-default' -o /dev/null -s http://localhost:8080/healthz

# Disabled: Alloy's running state is the enable/disable switch.
docker compose -f deployments/compose.yml stop alloy
docker compose -f deployments/compose.yml ps alloy
curl -H 'X-Request-ID: log-config-disabled' -o /dev/null -s -w '%{http_code}\n' http://localhost:8080/healthz
docker compose -f deployments/compose.yml logs backend

# Rejected protected destination: /dev/null exercises the read-only credential
# mount without creating or printing a secret; the non-push path returns an error.
LOKI_URL=http://loki:3100/rejected \
LOKI_TENANT_ID=acceptance-test \
LOKI_BEARER_TOKEN_FILE=/dev/null \
docker compose -f deployments/compose.yml up -d --force-recreate alloy
curl -H 'X-Request-ID: log-config-rejected' -o /dev/null -s http://localhost:8080/healthz
docker compose -f deployments/compose.yml logs alloy
docker compose -f deployments/compose.yml logs backend

# Restore the local default. The overrides above were exported into this
# shell, so they must be unset explicitly before the recreation.
unset LOKI_URL LOKI_TENANT_ID LOKI_BEARER_TOKEN_FILE
docker compose -f deployments/compose.yml up -d --force-recreate alloy
```

The default event must arrive in Loki. The disabled and rejected cases must return the normal backend status and leave backend stderr available. Rejected-delivery diagnostics may identify an HTTP or authorization failure but must contain no bearer-token value, request body, account identifier, or privacy sentinel.

## Validate an aggregation outage

Create a disposable acceptance account while the stack is healthy, retaining cookies in a temporary file outside the repository:

```bash
cookie_jar=$(mktemp)
email="log-outage-$(date +%s)@example.invalid"
username="log-outage-$(date +%s)"
curl -sS -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d "{\"first_name\":\"Outage\",\"last_name\":\"Tester\",\"username\":\"${username}\",\"email\":\"${email}\",\"password\":\"OutagePass123!\"}"
curl -sS -c "$cookie_jar" -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"${email}\",\"password\":\"OutagePass123!\"}"
```

Stop Loki but leave Alloy running so delivery failures remain observable. For 15 minutes, repeatedly exercise liveness, successful authentication, authenticated publication, and notification-history reads:

```bash
docker compose -f deployments/compose.yml stop loki
outage_end=$((SECONDS + 900))
sequence=0
while [ "$SECONDS" -lt "$outage_end" ]; do
  sequence=$((sequence + 1))
  curl -fsS -o /dev/null http://localhost:8080/healthz
  curl -fsS -o /dev/null -b "$cookie_jar" http://localhost:8080/notifications
  curl -fsS -o /dev/null -b "$cookie_jar" -X POST http://localhost:8080/posts \
    -H 'Content-Type: application/json' \
    -d "{\"title\":\"Outage ${sequence}\",\"content\":\"Aggregation outage acceptance event ${sequence}\",\"tags\":[\"acceptance\"]}"
  curl -fsS -o /dev/null -c "$cookie_jar" -X POST http://localhost:8080/auth/login \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"${email}\",\"password\":\"OutagePass123!\"}"
  sleep 30
done
docker compose -f deployments/compose.yml logs backend
docker compose -f deployments/compose.yml logs alloy
docker compose -f deployments/compose.yml start loki
curl -H 'X-Request-ID: log-outage-recovered' -o /dev/null -s http://localhost:8080/healthz
docker compose -f deployments/compose.yml logs alloy
rm -f "$cookie_jar"
```

Every workflow must retain its normal success status without waiting on Loki. Local stderr logs remain available during the outage, Alloy reports delivery failures without application data or credentials, and `log-outage-recovered` arrives after recovery. The initial release is best effort: events emitted while Loki is unavailable can be lost and are not guaranteed to replay.

## Automated checks

```bash
go test ./internal/adapters/logger ./internal/api/middleware ./cmd
make test
make check
make test-cover
go tool cover -func=test/coverage/coverage.out | tail -1
cd web
npm test -- --run
cd ..
PROJECT1_EMAIL='<acceptance-email>' PROJECT1_PASSWORD='<acceptance-password>' scripts/verify-request-latency.py
```

Compare the final backend coverage total with `specs/008-external-log-aggregation/coverage-baseline.txt`; it must not decrease. The latency verifier exercises representative authenticated read and write operations under normal local load and fails unless read p95 is below 200 ms and write p95 is below 500 ms. The implementation also adds focused tests for the safe completion-event schema, request correlation, and prohibited-field filtering. Refer to the [deployment contract](./contracts/log-aggregation-deployment.md) and [data model](./data-model.md) when a validation fails.

## Recorded acceptance results

Local Compose run on 2026-09-23 (`feat/loki-set-up`), Grafana Alloy v1.19.2, Loki 3.7.0, backend on port 8080. Loki is not host-published; stream queries below were issued from inside the Compose network.

| Check | Command | Observed result |
|---|---|---|
| Default destination (US2/AC1) | `curl -H 'X-Request-ID: log-config-default' .../healthz` | Event searchable in Loki with labels `app`/`environment`/`source` and safe fields only |
| Disabled collector (US2/AC2, FR-005) | `stop alloy` then `ps alloy` | Alloy absent from running services; `backend_status=200`; the event still appeared in `logs backend` |
| Rejected destination (US2/AC3) | recreate Alloy with `LOKI_URL=.../rejected`, `LOKI_TENANT_ID=acceptance-test`, `LOKI_BEARER_TOKEN_FILE=/dev/null` | `backend_status=200`; Alloy logged `final error sending batch, no retries left, dropping data` with `status=404` and no credential value |
| Restore default | `unset` the three overrides, recreate Alloy | `log-config-restored-2` delivered to Loki |
| 100 events, 20 contexts (SC-001, SC-002) | 100 requests using `log-context-01`…`log-context-20` | All 20 correlation contexts present in Loki; every event searchable immediately |
| Prohibited values (SC-004) | The sentinel requests above | `LOG_PRIVATE_SENTINEL`, `TOKEN_SENTINEL`, `COOKIE_SENTINEL`, `ACCOUNT_SENTINEL`, `PASSWORD_SENTINEL`, `PRIVATE_MESSAGE_SENTINEL`: 0 matches in Loki and 0 in `logs alloy` |
| Delivery hygiene | Query `message="request error"` records | Exported projection held only `timestamp`, `level`, `message`, `instance_id`, `request_id`, `method`, `status`, `error_code`; raw `path`, `user`, and error text stayed in backend stderr |
| Read/write p95 (Constitution IV) | `PROJECT1_EMAIL=... PROJECT1_PASSWORD=... scripts/verify-request-latency.py` | read p95 18.68 ms (< 200 ms), write p95 13.51 ms (< 500 ms) |
| Coverage (Constitution II) | `make test-cover` | 27.5% total, baseline 27.2% — no decrease |
| Quality gates | `make check`, `npm test -- --run` | both exit 0 |
| 15-minute destination outage (US3/AC1, SC-003) | `stop loki` at 08:38:43Z, then exercise liveness, authentication, publication, and notification reads every 30 s until 08:53:52Z | 30/30 iterations returned `healthz=200`, `notifications=200`, `post=201`, `login=200`; 0 workflow failures; backend stderr stayed available throughout |
| Delivery-failure observability (US3/AC2, FR-007) | `logs alloy` during the outage | 2 × `final error sending batch, no retries left, dropping data` (`status:-1`, `dial tcp: lookup loki … no such host`), the first at 08:45:49Z, about seven minutes in once the writer's retry backoff was exhausted; no credentials or user content. The delay means the first failure report can lag the outage by several minutes |
| Sentinels during the outage | `logs alloy` filtered for the matrix values | 0 matches |
| Destination recovery (edge case, US3 checkpoint) | `start loki`, then emit a marker after recovery | The post-recovery marker was delivered within about 20 s with no user action. A second marker emitted during the outage was also delivered once Loki returned, because Alloy's in-memory queue still held it while the outage was shorter than its retry budget |
| Best-effort boundary | Outage long enough to exhaust the writer's retries | The writer reports `final error sending batch, no retries left, dropping data` and those entries are dropped, never replayed; `log-config-disabled` (collector stopped) and `log-config-rejected` (non-retryable 404) produced 0 entries in Loki |
| Tenant identity not exposed (T024) | Recreate Alloy with `LOKI_TENANT_ID=acceptance-test` against the rejected destination | Delivery failures reported as before, `"tenant":""`, and 0 occurrences of `acceptance-test` anywhere in `logs alloy`; delivery still worked for the default and restored configurations |

Queries that use a parser (`| json`) return the parsed fields inside the response's `stream` object, which can look like extra stored labels. The stored label set is exactly the three allowed labels; verify it with `GET /loki/api/v1/labels`, which reports only `app`, `environment`, and `source`.

Tenant handling: Alloy echoes a configured `tenant_id` in its own delivery-failure diagnostics, which would contradict the data model's rule that the tenant identity is never logged. The pipeline therefore sends the tenant as the standard `X-Scope-OrgID` header instead of the `tenant_id` argument, which the composer supplies from `LOKI_TENANT_ID`; the header is not echoed, so the tenant never appears in `logs alloy`. An empty header value is sent when no tenant is configured, which Loki ignores while multi-tenancy is disabled.

