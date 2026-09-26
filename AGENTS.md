# AI Development Instructions

## Context & Initialization

- Primary Reference: Read @README.md first to understand the project architecture, dependencies, and core objectives. Do not execute further commands until the project scope is clear.

## Architecture & Ownership

- This is a Go/Echo backend and Next.js App Router frontend. PostgreSQL stores application data; RabbitMQ carries notification events; browser notifications use SSE.
- Preserve Clean Architecture boundaries. Each bounded context in `internal/` owns its domain, ports, use cases, adapters, HTTP API, and context-specific configuration:
  - `identity`: accounts, sessions, authentication, and user search
  - `publishing`: posts, comments, and likes
  - `social`: follows and personal feeds
  - `notifications`: notification persistence, broker events, and SSE delivery
  - `featureflags`: flag administration and evaluation
  - `operations`: readiness and HTTP metrics
- Keep `internal/platform` small and technical only: bootstrap configuration, logging, validation, broker clients, shared pagination/problem vocabulary, common HTTP error handling, and authenticated-principal extraction. Do not place product business rules there.
- Keep cross-context dependencies explicit and narrow. Depend on ports or event contracts rather than another context's repository or adapter.
- Generated GoMock implementations live in `internal/testkit/mocks`. Run `make mocks` whenever a port interface changes, and include resulting generated changes in the same commit.

## HTTP API & Frontend Contract

- Keep route registration inside the owning context's `internal/<context>/api/routes.go`. `cmd/main.go` is composition/bootstrap only: it configures global middleware, Swagger, infrastructure, and calls each context's `RegisterRoutes` function.
- Preserve existing route paths, methods, middleware ordering, authorization, and feature-flag gates unless the request explicitly changes the public API. Route ownership must not change observable behavior.
- Use the existing RFC 9457 problem-details error handling. Do not return ad-hoc error response shapes from handlers.
- The frontend is a same-origin BFF: browser code calls `/api/*`, while Next.js route handlers forward cookies to the Go API using server-only `API_URL`. When changing an API route, update its `web/app/api/` proxy and relevant frontend consumers/tests together.
- Keep live notifications on `GET /notifications/stream` and its BFF proxy at `/api/notifications/stream`; preserve cookie forwarding and SSE semantics.
- Update checked-in Swagger artifacts with `make docs` whenever handler annotations or the OpenAPI contract change.

## Configuration, Observability & Deployments

- Configuration is loaded through Viper. Environment overrides use uppercase underscore-separated keys, for example `DATABASE_HOST`, `JWT_PRIVATE_KEY_PATH`, and `MESSAGE_BROKER_RABBITMQ_URL`.
- Keep `/healthz` and `/status` unauthenticated and non-sensitive. `/healthz` must not contact dependencies; `/status` is the dependency-readiness probe used by Compose.
- `/metrics` must remain protected by the dedicated monitoring credential. Never add unbounded/sensitive metric labels (account IDs, tokens, session values, raw paths, query strings, request bodies, or raw dependency errors).
- Never commit JWT keys, monitoring passwords, Loki bearer tokens, or other secrets. Use file-backed secrets; do not put token contents in environment variables.
- Before changing deployment configuration, read `deployments/README.md`. Preserve the Compose health probe at `/status`, Prometheus scrape authentication for `/metrics`, and the frontend's in-network `API_URL=http://backend:8080` unless the request explicitly changes them.
- Validate Compose changes without starting services using `GRAFANA_ADMIN_PASSWORD='<temporary-value>' docker compose -f deployments/compose.yml config -q`. Starting, stopping, recreating, or deleting deployment services/volumes requires explicit user authorization.

## Verification & Delivery

- Run focused tests for changed packages, then run `go test ./...` for backend changes. When the default Go cache is unavailable, use a writable temporary cache such as `GOCACHE=/private/tmp/project-one-go-build-cache`.
- For frontend changes, run the applicable checks from `web/`: `npm run lint`, `npm test -- --run`, and `npm run build` as appropriate.
- Run `gofmt` on changed Go files and `git diff --check` before delivery. Do not overwrite or discard unrelated working-tree changes.
- Use commit messages in the form `<type>(<scope-or-ticket>): <description>`, with one of `feat`, `fix`, `chore`, `refactor`, `docs`, or `test`.
