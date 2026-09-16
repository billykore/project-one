# Project One

Project One is a full-stack social publishing application. It combines a Go/Echo API organized with Clean Architecture, a Next.js App Router frontend, PostgreSQL persistence, RabbitMQ-backed notification events, and live browser updates over Server-Sent Events (SSE).

## Features

- Registration, login, logout, and RSA-signed JWT sessions stored in HTTP-only cookies
- User profiles, profile editing, password changes, user search, and follow relationships
- Post creation, browsing, editing, deletion, comments, and idempotent likes
- A cursor-paginated personal feed with infinite scrolling
- Persistent follow, like, and comment notifications delivered live over SSE
- RFC 9457 Problem Details error responses with request IDs
- Swagger/OpenAPI documentation and generated Postman API tests
- Backend and frontend unit tests, linting, and CI workflows

## Stack

### Backend

- Go 1.26.2
- Echo 4.15, GORM, and PostgreSQL
- RabbitMQ for notification events; Kafka and in-memory adapters are also present
- Viper configuration, `log/slog` structured logging, Validator v10, bcrypt, and JWT v5
- Testify, GoMock, and Swaggo

### Frontend

- Next.js 16.2.9 with the App Router
- React 19.2.4 and TypeScript 5
- Tailwind CSS 4
- Vitest and Playwright tooling

## Architecture

The backend keeps business rules independent of delivery and infrastructure concerns:

- `internal/core/domain`: entities, value objects, and domain errors
- `internal/core/ports`: interfaces used at architectural boundaries
- `internal/core/usecase`: application business logic
- `internal/adapters`: PostgreSQL repositories, token/password services, pub/sub clients, logging, validation, and SSE connection management
- `internal/api`: Echo handlers, DTOs, and middleware

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

This starts PostgreSQL 17, RabbitMQ 4, the Go API, and the Next.js frontend. You need Docker with Compose and OpenSSL installed.

1. Generate the local RSA key pair used to sign JWTs:

   ```bash
   mkdir -p configs/keys
   openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out configs/keys/jwt-private.pem
   openssl pkey -in configs/keys/jwt-private.pem -pubout -out configs/keys/jwt-public.pem
   ```

2. Build and start the stack:

   ```bash
   make compose-up
   ```

3. Open the services:

   - Frontend: <http://localhost:3000>
   - API health check: <http://localhost:8080/status>
   - Swagger UI: <http://localhost:8080/swagger/index.html>

4. Stop the stack when finished:

   ```bash
   make compose-down
   ```

On the first start of a new PostgreSQL volume, the container applies every `db/migrations/*.up.sql` file. For an existing volume, apply new migrations explicitly with `make migrate-up`; the initialization script does not rerun.

See [deployments/README.md](deployments/README.md) for service configuration, lifecycle commands, and data-volume behavior.

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
| `make build` | Build the backend binary at `bin/main` |
| `make run` | Build and run the backend |
| `make test` | Run backend tests with the race detector |
| `make test-cover` | Run backend tests and write the HTML coverage report |
| `make mocks` | Regenerate GoMock implementations for all core ports |
| `make vet` | Run `go vet` |
| `make lint` | Run `golangci-lint` |
| `make docs` | Format Swagger annotations and regenerate `api/swagger` |
| `make check` | Run docs generation, vet, lint, and backend tests |
| `make migrate-create name=...` | Create a numbered up/down migration pair |
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
