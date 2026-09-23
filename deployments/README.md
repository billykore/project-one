# Docker Compose deployment

`compose.yml` defines the complete local Project One stack. It is intended for development and integration testing rather than as a hardened production deployment.

## Services

| Service | Image/build | Published port | Purpose |
| :--- | :--- | :--- | :--- |
| `postgres` | `postgres:17-alpine` | `5432` | Application data |
| `rabbitmq` | `rabbitmq:4-alpine` | `5672` | Notification event transport |
| `keys` | `alpine:3.22` | None | Copies JWT keys into a protected shared volume |
| `backend` | Root `Dockerfile` | `8080` | Go API and notification consumer; health-checked through `/status` |
| `frontend` | `web/Dockerfile` | `3000` | Next.js standalone server |
| `prometheus` | `prom/prometheus:v3.7.3` | `9090` | Scrapes the API `/metrics` endpoint and stores measurements |
| `loki` | `grafana/loki:3.7.0` | None | Private, seven-day local log store |
| `alloy` | `grafana/alloy:v1.19.2` | None | Filters and forwards backend container logs |
| `grafana` | `grafana/grafana:12.2.0` | `3001` | Serves metrics dashboards and Loki Explore |

PostgreSQL, RabbitMQ, and the backend have health checks. The backend starts only after the database and broker are healthy and JWT key preparation has completed; Prometheus starts after the backend container is created and Grafana after Prometheus. The frontend starts after the backend container is created. The backend probe uses Alpine's built-in `wget` to request `/status`.

## Prerequisites

- Docker with Compose support
- An RSA private/public key pair at `configs/keys/jwt-private.pem` and `configs/keys/jwt-public.pem`
- A local monitoring password secret and a Grafana administrator password (see [Observability](#observability))

Generate development keys from the repository root:

```bash
mkdir -p configs/keys
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out configs/keys/jwt-private.pem
openssl pkey -in configs/keys/jwt-private.pem -pubout -out configs/keys/jwt-public.pem
```

The key directory is ignored by Git. Do not commit private keys.

## Observability

### Monitoring secret

The backend exposes bounded operational metrics on `/metrics`. Prometheus authenticates with a dedicated monitoring credential that is deliberately separate from user authentication and never appears in source control, configuration files, logs, or metric labels.

Create the local secret before the first `make compose-up`:

```bash
mkdir -p deployments/observability/secrets
printf '%s' '<monitoring-password>' > deployments/observability/secrets/metrics-password
# The Prometheus image runs as an unprivileged user, so the file must stay
# world-readable. The directory is ignored by Git; do not commit the secret.
chmod 644 deployments/observability/secrets/metrics-password
```

The directory is ignored by Git. The same file is mounted into the backend as the `metrics-password` Compose secret (`/run/secrets/metrics-password`) and into Prometheus at `/etc/prometheus/metrics-password`, which `prometheus.yml` reads through `password_file`. The username is the fixed machine identity `projectone-metrics`, set on the backend through `MONITORING_USERNAME` and in the scrape job through `basic_auth.username`.

If `MONITORING_USERNAME` and `MONITORING_PASSWORD_FILE` are unset, or the referenced file is missing or empty, `/metrics` returns `401` for every request and startup fails when only one of the two settings is present.

Set the Grafana administrator password in the invoking shell, because the Compose file requires it:

```bash
export GRAFANA_ADMIN_PASSWORD='<local-grafana-password>'
```

### Health probes

| Route | Authentication | Meaning |
| :--- | :--- | :--- |
| `/healthz` | None | Process liveness only. Never contacts a dependency. |
| `/status` | None | Readiness report for the backend, PostgreSQL, and the notification subscriber. Returns `503` and identifies the affected component when any required dependency is not up. |

The `/status` body stays non-sensitive and holds no address, credential, token, account identifier, or raw dependency error, so it remains safe as the container health probe. Callers that need process liveness rather than readiness must use `/healthz`.

### Collecting and viewing measurements

Scrape the API manually with the monitoring credential:

```bash
curl -u projectone-metrics:'<monitoring-password>' http://localhost:8080/metrics
curl -o /dev/null -s -w '%{http_code}\n' http://localhost:8080/metrics   # 401 without credentials
```

Prometheus is at <http://localhost:9090>, where `/targets` shows the `project-one-api` job. Grafana is at <http://localhost:3001>; sign in as `admin` with `GRAFANA_ADMIN_PASSWORD` and open the provisioned **Project One Health** dashboard, which shows readiness, dependency state, request rate, 4xx and 5xx rates, request duration p95, and instance start time. Datasources and dashboards are provisioned from `grafana/provisioning` and `grafana/dashboards`, so no manual setup is required.

### Aggregated application logs

Alloy discovers only the Compose `backend` service, parses its JSON records, rebuilds each outgoing event from a safe field allowlist, applies secret filtering, and sends it to Loki. Loki is private to the Compose network and has no host-published port. The backend writes only to stderr and neither starts nor waits on Loki or Alloy.

In Grafana, open **Explore**, select **Loki**, and query:

```logql
{app="project-one", environment="development", source="backend"} | json
```

The stream has exactly the bounded labels `app`, `environment`, and `source`. Correlate events with the JSON `request_id` field rather than adding it as a label. Local Loki data is retained for seven days; durable replay and long-term archival are outside this deployment's scope.

Alloy running is the aggregation enablement switch. Stop it to disable forwarding while preserving backend stderr:

```bash
docker compose -f deployments/compose.yml stop alloy
docker compose -f deployments/compose.yml logs backend
docker compose -f deployments/compose.yml start alloy
```

For a protected external Loki-compatible destination, set its full push URL and optional tenant. Store the bearer token in an ignored local file and pass only its path:

```bash
export LOKI_URL='https://logs.example.invalid/loki/api/v1/push'
export LOKI_TENANT_ID='<tenant>'
export LOKI_BEARER_TOKEN_FILE='/absolute/path/to/ignored/loki-token'
docker compose -f deployments/compose.yml up -d --force-recreate alloy
```

Compose mounts the credential read-only at `/run/secrets/loki-bearer-token`; its contents are never placed in Compose environment values or application artifacts. The tenant is sent as the standard `X-Scope-OrgID` header rather than Alloy's `tenant_id` argument, because Alloy echoes `tenant_id` in its own delivery-failure diagnostics; with this wiring the tenant never appears in `logs alloy`. Unset all three overrides and recreate Alloy to restore private local Loki.

If events are missing, first check `docker compose -f deployments/compose.yml ps loki alloy`, then inspect `docker compose -f deployments/compose.yml logs alloy`. A delivery or authorization error indicates a destination, tenant, network, or credential problem; it does not affect backend availability. Confirm backend stderr independently with `docker compose -f deployments/compose.yml logs backend`. Do not paste credentials or user data into diagnostic searches.

`deployments/prometheus.yml` holds the scrape job, `deployments/grafana/provisioning/datasources/project-one-prometheus.yml` the datasource, `deployments/grafana/provisioning/dashboards/project-one-health.yml` the dashboard provider, and `deployments/grafana/dashboards/project-one-health.json` the versioned dashboard. Keep the datasource file's current name: editors apply the Prometheus scrape-configuration schema to any file named exactly `prometheus.yml`, which would flag `apiVersion` and `datasources` as invalid even though Grafana accepts them.

## Lifecycle

Run these commands from the repository root:

```bash
make compose-up       # build images and start in the background
make compose-stop     # stop containers, preserving them
make compose-start    # restart existing stopped containers
make compose-down     # remove containers and network, retaining volumes
```

To inspect the stack directly:

```bash
docker compose -f deployments/compose.yml ps
docker compose -f deployments/compose.yml logs -f backend frontend prometheus grafana
```

The frontend is served at <http://localhost:3000>, the API at <http://localhost:8080>, Prometheus at <http://localhost:9090>, and Grafana at <http://localhost:3001>. In non-production mode, Swagger UI is at <http://localhost:8080/swagger/index.html>.

## Configuration

The Compose file supplies backend configuration through environment variables. Override database and RabbitMQ credentials in the invoking shell or in a root `.env` file read by Docker Compose:

```dotenv
POSTGRES_DB=project1
POSTGRES_USER=project1
POSTGRES_PASSWORD=change-me
RABBITMQ_USER=project1
RABBITMQ_PASSWORD=change-me
```

These variables configure both the infrastructure containers and the backend connection strings. The frontend receives `API_URL=http://backend:8080`, which is reachable only within the Compose network.

## Database initialization and volumes

`init-db.sh` runs every `db/migrations/*.up.sql` file, in filename order, when PostgreSQL initializes a new data volume. PostgreSQL only runs scripts under `/docker-entrypoint-initdb.d` for an empty data directory.

For schema changes against an existing `postgres-data` volume, run the migration CLI explicitly. Replace the placeholders below with the same database name, user, and password that initialized the volume (values in a Compose `.env` file are not automatically exported to your shell):

```bash
make migrate-up dsn='postgres://<user>:<password>@localhost:5432/<database>?sslmode=disable'
```

The stack uses seven named volumes:

- `postgres-data` for PostgreSQL data
- `rabbitmq-data` for broker data
- `jwt-keys` for the runtime copy of the RSA keys
- `prometheus-data` for collected measurements
- `grafana-data` for Grafana state
- `loki-data` for the local log store
- `alloy-data` for collector component state (not a durable replay guarantee)

`make compose-down` preserves these volumes. Running `docker compose -f deployments/compose.yml down -v` deletes all seven volumes and permanently removes their local data.

## Container images

The backend image uses a Go 1.26.2 Alpine build stage and an Alpine runtime with BusyBox `wget` for its health probe; it executes as UID/GID 65532. The frontend image uses Node.js 24 Alpine, Next.js standalone output, and the unprivileged `node` user.
