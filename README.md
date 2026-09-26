# Project One

Project One is a full-stack social publishing application. It combines a Go/Echo API organized with Clean Architecture, a Next.js App Router frontend, PostgreSQL persistence, RabbitMQ-backed notification events, and live browser updates over Server-Sent Events (SSE).

## Features

- Registration, login, logout, and RSA-signed JWT sessions stored in HTTP-only cookies
- User profiles, profile editing, password changes, user search, and follow relationships
- Post creation, browsing, editing, deletion, comments, and idempotent likes
- A cursor-paginated personal feed with infinite scrolling
- Persistent follow, like, and comment notifications delivered live over SSE
- RFC 9457 Problem Details error responses with request IDs
- Separate liveness (`/healthz`) and readiness (`/status`) probes with per-dependency state
- Bounded-label Prometheus metrics on an authenticated `/metrics` endpoint, plus a provisioned Grafana dashboard
- Swagger/OpenAPI documentation and generated Postman API tests
- Backend and frontend unit tests, linting, and CI workflows

## Stack

### Backend

- Go 1.26.2
- Echo 4.15, GORM, and PostgreSQL
- RabbitMQ for notification events; Kafka and in-memory adapters are also present
- Viper configuration, `log/slog` structured logging, Validator v10, bcrypt, and JWT v5
- Official Prometheus Go client with a dedicated registry, bounded labels, and standard Go/process collectors
- Testify, GoMock, and Swaggo

### Frontend

- Next.js 16.2.9 with the App Router
- React 19.2.4 and TypeScript 5
- Tailwind CSS 4
- Vitest and Playwright tooling

## Architecture

The backend keeps business rules independent of delivery and infrastructure concerns, and assigns each business capability to a bounded context:

- `internal/identity`: accounts, credentials, sessions, authentication, and user search
- `internal/publishing`: posts, comments, and likes
- `internal/social`: follows and the personal feed
- `internal/notifications`: notification persistence and the broker event contract
- `internal/featureflags`: flag administration and evaluation
- `internal/operations`: readiness assessment and HTTP metrics contracts
- `internal/platform`: the deliberately small shared kernel—pagination, transport-neutral problem vocabulary, and technical ports for logging, messaging, and validation

Every context owns its domain, ports, use cases, adapters, HTTP API, and any runtime configuration it needs. `platform` retains only bootstrap/technical concerns: database and broker configuration, logging, validation, broker clients, common HTTP error handling, and authenticated-principal extraction. Cross-context dependencies are explicit and narrow (for example, publishing emits the notifications event contract). Generated mocks live in `internal/testkit/mocks`.

The frontend uses Next.js route handlers as a same-origin backend-for-frontend (BFF). Browser requests carry the session cookies to `/api/*`; the route handlers forward them to the Go API through the server-only `API_URL` setting.

```mermaid
flowchart LR
    browser[Browser]

    subgraph next[Next.js frontend]
        pages[App Router pages and components]
        bff[Route handlers /api/*]
        sseClient[EventSource client]
    end

    subgraph go[Go API]
        handlers[Echo handlers and middleware]
        usecases[Use cases]
        ports[Ports]
        adapters[Adapters]
        sse[SSE manager]
    end

    postgres[(PostgreSQL)]
    rabbit[(RabbitMQ)]

    browser --> pages
    pages --> bff
    sseClient -->|GET /api/notifications/stream| bff
    bff -->|HTTP and cookies| handlers
    handlers --> usecases
    usecases --> ports
    ports --> adapters
    adapters --> postgres
    usecases -->|publish| rabbit
    rabbit -->|consume| handlers
    handlers --> sse
    sse -->|text/event-stream| bff
```

## Quick start with Docker Compose

This starts PostgreSQL 17, RabbitMQ 4, the Go API, the Next.js frontend, Prometheus, and Grafana. You need Docker with Compose and OpenSSL installed.

1. Generate the local RSA key pair used to sign JWTs:

   ```bash
   mkdir -p configs/keys
   openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out configs/keys/jwt-private.pem
   openssl pkey -in configs/keys/jwt-private.pem -pubout -out configs/keys/jwt-public.pem
   ```

2. Create the local monitoring secret and Grafana administrator password. Both are ignored by Git:

   ```bash
   mkdir -p deployments/observability/secrets
   printf '%s' '<monitoring-password>' > deployments/observability/secrets/metrics-password
   export GRAFANA_ADMIN_PASSWORD='<local-grafana-password>'
   ```

3. Build and start the stack:

   ```bash
   make compose-up
   ```

4. Open the services:

   - Frontend: <http://localhost:3000>
   - Liveness probe: <http://localhost:8080/healthz>
   - Readiness probe: <http://localhost:8080/status>
   - Swagger UI: <http://localhost:8080/swagger/index.html>
   - Prometheus: <http://localhost:9090>
   - Grafana (Project One Health and Project One Logs dashboards): <http://localhost:3001>

5. Stop the stack when finished:

   ```bash
   make compose-down
   ```

On the first start of a new PostgreSQL volume, the container applies every `db/migrations/*.up.sql` file. For an existing volume, apply new migrations explicitly with `make migrate-up`; the initialization script does not rerun.

See [deployments/README.md](deployments/README.md) for service configuration, lifecycle commands, data-volume behavior, and the observability setup.

## Health, metrics, and logs

| Route | Authentication | Purpose |
| :--- | :--- | :--- |
| `/healthz` | None | Process liveness. Never contacts a dependency, so a starting instance is live before it is ready. |
| `/status` | None | Readiness report for PostgreSQL and the notification subscriber. Returns `503` and identifies only the affected component when a required dependency is down, unknown, or cannot be assessed in time. |
| `/metrics` | Monitoring credential | Prometheus text exposition of request counts, request duration, readiness, dependency state, and process start time. |

`/healthz` and `/status` stay non-sensitive and unauthenticated so the deployment probe keeps working; `/metrics` requires the dedicated monitoring credential read from a deployment secret. Metrics carry only bounded labels: the HTTP method, the resolved route template, the status class, and the dependency name. No account identifier, token, session value, request body, query string, or raw dependency error is ever used as a label or returned in a health report.

Set `MONITORING_USERNAME` and `MONITORING_PASSWORD_FILE` to enable scraping; when they are unset, `/metrics` rejects every request. Prometheus scrapes `backend:8080/metrics` and Grafana reads it through the provisioned datasource. Alert rules, long-term storage, tracing, and automated remediation are intentionally out of scope.

The Compose deployment also runs a private Loki service and Grafana Alloy collector. Alloy reads only backend container output, forwards a safe structured allowlist, and leaves the API lifecycle independent of log delivery. The provisioned **Project One Logs** dashboard shows counts, level trends, and recent events. In Grafana Explore, select the provisioned **Loki** datasource and start with:

```logql
{app="project-one", environment="development", source="backend"} | json
```

Use `request_id` from the parsed JSON to correlate requests. Loki is intentionally not published on a host port. External Loki-compatible destinations use `LOKI_URL`, optional `LOKI_TENANT_ID`, and a read-only file path in `LOKI_BEARER_TOKEN_FILE`; never put bearer-token contents in environment variables or committed files. See [deployments/README.md](deployments/README.md#aggregated-application-logs) for setup, disabling, retention, privacy, and troubleshooting details.

## Local development

### Prerequisites

- Go 1.26.2
- Node.js 24 and npm (Node.js 20 is also used by the lint CI job)
- PostgreSQL and RabbitMQ
- OpenSSL
- Optional command-line tools: `migrate`, `swag`, and `golangci-lint`
- Python 3 and pip only when using the seed commands

### Backend

1. Copy the example configuration:

   ```bash
   cp configs/config.yaml.example configs/config.yaml
   ```

2. Generate `configs/keys/jwt-private.pem` and `configs/keys/jwt-public.pem` using the commands in the Docker quick start, then update database and RabbitMQ values in `configs/config.yaml`.

3. Apply the migrations:

   ```bash
   make migrate-up dsn="postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
   ```

4. Start the API at <http://localhost:8080>:

   ```bash
   make run
   ```

The configuration file path can be changed with `make run config=/path/to/config-dir`. Bound settings can also be overridden by uppercase, underscore-separated environment variables—for example, `DATABASE_HOST`, `JWT_PRIVATE_KEY_PATH`, or `MESSAGE_BROKER_RABBITMQ_URL`.

Scraping is optional locally: leave `monitoring.username` and `monitoring.password_file` unset and `/metrics` rejects every request. To enable it, set both (or `MONITORING_USERNAME` and `MONITORING_PASSWORD_FILE`) and point the password file at an existing, non-empty secret file. `configs/config.yaml.example` documents the keys.

### Frontend

In a second terminal:

```bash
cd web
npm ci
API_URL=http://localhost:8080 npm run dev
```

The frontend is available at <http://localhost:3000>. `API_URL` is server-only and defaults to `http://localhost:8080`; it is not exposed to browser JavaScript.

See [web/README.md](web/README.md) for frontend routes, rendering boundaries, API proxying, and tests.

## Developer commands

| Command | Description |
| :--- | :--- |
| `make help` | List documented Make targets |
| `make build` | Build the backend binary at `build/bin/main` |
| `make run` | Build and run the backend |
| `make test` | Run backend tests with the race detector |
| `make test-cover` | Run backend tests and write the HTML coverage report |
| `make mocks` | Regenerate GoMock implementations for all context and platform ports |
| `make vet` | Run `go vet` |
| `make lint` | Run `golangci-lint` |
| `make docs` | Format Swagger annotations and regenerate `api/swagger` |
| `make check` | Run docs generation, vet, lint, and backend tests |
| `make migration-create name=...` | Create a numbered up/down migration pair |
| `make migrate-up dsn=...` | Apply migrations; optionally pass `steps=N` |
| `make migrate-down dsn=...` | Revert migrations; optionally pass `steps=N` |
| `make seed-users dsn=...` | Install seed dependencies and insert 20 generated users |
| `make seed-posts dsn=...` | Install seed dependencies and insert 100,000 generated posts |
| `make compose-up` | Build and start the Compose stack |
| `make compose-down` | Stop the stack and remove its containers and network |
| `make compose-start` | Start existing stopped Compose containers |
| `make compose-stop` | Stop Compose containers without removing them |
| `make githooks` | Activate the repository's local Git hooks |
| `make clean` | Remove backend build artifacts |

The seed commands use the provided DSN as `DATABASE_URL`. Without one, the scripts fall back to `postgresql://postgres:postgres@localhost:5432/my_go_db`. Seeded rows contain generated fixture data and are intended for development datasets. Set `POST_COUNT` or `BATCH_SIZE` to override the post seeder defaults.

Frontend checks run from `web/`:

```bash
npm run lint
npm test -- --run
npm run build
```

## API and live notifications

Swagger artifacts are checked in under `api/swagger/`. Run `make docs` after changing handler annotations. Swagger UI is served at `/swagger/index.html` unless `APP_ENV=production`.

The notification flow is:

1. Follow, like, and comment use cases publish an event to RabbitMQ.
2. The notification consumer persists the event in PostgreSQL.
3. Connected clients receive the new notification from `GET /notifications/stream` as SSE.
4. Clients retrieve notification history from `GET /notifications` and can mark one or all notifications as read.

The SSE endpoint accepts the same `access_token` cookie or `Authorization: Bearer <token>` header as other protected endpoints. The browser connects through the frontend route at `/api/notifications/stream`, which forwards its cookies to the backend. Streams include a keepalive comment every 30 seconds.

## Project structure

```text
├── api/swagger/          # Generated OpenAPI documentation
├── build/                # Packaging and CI placeholders
├── cmd/                  # Backend entry point and RSA key loading
├── configs/              # YAML configuration and local JWT keys
├── db/migrations/        # Versioned PostgreSQL migrations
├── db/seeds/             # Development user seeding utility
├── deployments/          # Docker Compose stack and DB initialization
├── docs/                 # Feature requirements, designs, and plans
├── githooks/             # Repository Git hooks
├── internal/             # Backend domain, ports, use cases, adapters, and API
├── scripts/              # Postman collection generator
├── specs/                # Spec Kit feature artifacts
├── test/                 # API collections, reports, screenshots, and coverage
└── web/                  # Next.js frontend
```

## Commit convention

Commit messages use:

```text
<type>(<scope-or-ticket>): <description>
```

Valid types are `feat`, `fix`, `chore`, `refactor`, `docs`, and `test`. Run `make githooks` to enable the checks described in [githooks/README.md](githooks/README.md).

## License

See [LICENSE.md](LICENSE.md).
