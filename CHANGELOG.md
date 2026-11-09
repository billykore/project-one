# Changelog

All notable changes to Project One are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- The notification outbox parks an event as failed after 60 delivery attempts
  (about an hour of broker failure) instead of retrying it forever, and an
  hourly cleanup deletes published events after 7 days. Migration `000027`
  adds the `failed_at` column; apply it to existing databases with
  `make migrate-up`. `deployments/README.md` explains how to requeue parked
  events.
- `projectone_outbox_pending_messages`,
  `projectone_outbox_oldest_pending_age_seconds`, and
  `projectone_outbox_failed_messages` gauges on `/metrics`, with matching
  panels on the Project One Health dashboard.

### Changed

- The social feed no longer uses publishing's `Post` aggregate. Social reads
  its own `TimelinePost` projection through a `TimelineReader` adapter, and a
  `depguard` rule keeps `publishing/domain` out of other contexts' domain,
  ports, use cases, and API packages. The `GET /feeds` response is unchanged.

### Fixed

- `make docs` (and therefore `make check`) generates the Swagger artifacts
  again. Three contexts each declared a `dto.MessageResponse`, which `swag`
  could not resolve. They are now `IdentityMessageResponse`,
  `NotificationMessageResponse`, and `PublishingMessageResponse`; response
  bodies are unchanged, only the OpenAPI schema names differ.
- `message_broker.type: kafka` was accepted although the application always
  connects to RabbitMQ, so such a configuration silently used the RabbitMQ
  settings and reported its broker health as `kafka`. Startup now fails
  unless the type is `rabbitmq`.

### Removed

- The unused `PostResponse` and `PostsListResponse` DTOs from the identity
  context, which duplicated publishing's type names.
- The Kafka and in-memory pub/sub adapters, the broker factory, the
  `message_broker.kafka.*` settings, and the `segmentio/kafka-go` dependency.
  None of them could be selected at runtime: the application always wires
  RabbitMQ.

## [5.2.0] - 2026-11-04 (035db71)

### Added

- A transactional notification outbox (migration `000026`). Follow, like, and
  comment events are written in the same database transaction as the change
  that caused them and dispatched to the broker by a background worker. Apply
  the migration to existing databases with `make migrate-up`.
- Project development rules for contributors and AI agents in `AGENTS.md`.

### Changed

- Moved HTTP route registration into each bounded context's `api/routes.go`;
  `cmd/main.go` now only composes the application. Routes, middleware, and
  feature-flag gates are unchanged.
- Grouped adapter packages by role (for example `adapters/repository` and
  `adapters/evaluator`).
- Authenticated handlers read a transport-neutral principal from the request
  context instead of the identity user.
- The Compose backend service runs a prebuilt image instead of building from
  the `Dockerfile`.
- Renamed the restricted shared technical kernel from `internal/platform` to
  `internal/pkg`. Branches importing `internal/platform/...` must update their
  import paths; regenerate mocks with `make mocks`.
- Enforced the shared-kernel boundary with a `depguard` lint rule: `internal/pkg`
  can no longer import a bounded context.
- Moved context-owned sentinel errors from `internal/pkg/problem` into each
  context's `domain` package; each context registers its problem-details
  mappings from its `RegisterRoutes`. HTTP status codes and messages are unchanged.
- Replaced the `auth.Principal` alias of the identity user with a struct that
  carries only `ID` and `Username`.
- Moved aggregate configuration loading into `cmd`; each context validates its
  own settings.
- Moved the notification broker consumer out of the HTTP handler into
  `notifications/adapters/consumer`; the use case now owns event recording and
  realtime delivery through a `RealtimeNotifier` port.
- Publishing, social, and notifications no longer use identity's `User`
  aggregate. Each defines its own `ID`/`Username` projection (`Author`,
  `Account`, `Actor`) filled by a lookup adapter, and a `depguard` rule keeps
  `identity/domain` out of their domain, ports, use cases, and API packages.
- Removed the unused identity use case and validator dependencies from the
  notification HTTP handler.

### Fixed

- Redelivered notification events are acknowledged as duplicates again. With
  GORM's `TranslateError` enabled, the consumer's error-message check never
  matched, so each duplicate was retried and then dead-lettered.

## [5.1.0] - 2026-09-26 (fb60339)

### Changed

- Added version, revision, and build-date metadata to backend container images
  and tagged builds with both the release version and commit SHA.

## [5.0.0] - 2026-09-26 (00bdd0f)

### Added

- Database migrations that repair identity references and store follow, comment,
  like, and session relationships by stable user IDs.

### Changed

- Reorganized the backend as a modular monolith with bounded contexts for
  identity, publishing, social, notifications, feature flags, and operations;
  retained shared technical concerns in the restricted `platform` kernel.
- Updated authentication, authorization, session, post, comment, like, follow,
  feed, and notification flows to use authenticated user IDs rather than
  usernames or client-derived identities.
- Excluded generated coverage artifacts from version control.
- Reduced structured-log noise by skipping routine readiness and metrics
  requests.

### Fixed

- Made frontend server-side API fetches resolve to the internal API origin
  instead of the incoming request host.

## [4.1.0] - 2026-09-24 (7cd0b2c)

### Added

- External application log aggregation through Loki and Grafana Alloy.
- Provisioned Project One Logs Grafana dashboard and Loki datasource.

### Changed

- Added structured request logging with safe field forwarding and optional external Loki configuration.

## [4.0.0] - 2026-09-22 (5e19658)

### Added

- Separate liveness and readiness health probes with per-dependency status reporting.
- Authenticated Prometheus metrics with bounded labels for request and dependency observability.
- Prometheus and Grafana deployment configuration, including a provisioned Project One Health dashboard.

### Changed

- Updated configuration, API documentation, and deployment documentation for health checks and monitoring.

## [3.2.0] - 2026-09-18 (82c3fa4)

### Added

- Post CQRS specification, design artifacts, and a benchmark script for comparing post operation paths.
- `seed-posts` command that generates bulk post data for development.
- Playwright end-to-end login test, wired into the web CI workflow.
- Cursor value-object tests and pagination repository coverage.

### Changed

- Split post operations into separate command and query paths (CQRS), replacing the single post use case, handler, and repository with dedicated command and query implementations.
- Standardized API response handling across posts, users, followers, following, and notifications, including cursor-paginated response types and forwarded query parameters in frontend proxy routes.
- Updated Swagger definitions, the Postman collection, and API test-case documentation to match the new response structures.

## [3.1.0] - 2026-09-12 (a48663a)

### Changed

- Simplified `PostUseCase` initialization and removed its deprecated method.
- Corrected the feature-flag name used by the `CreatePost` test.
- Standardized Docker image tags on the first seven characters of the commit hash.
- Create CHANGELOG.md

## [3.0.0] - 2026-09-08 (775a29e)

### Added

- Feature-flag use cases with CRUD operations and audit logging.
- Database migrations for feature-flag settings, overrides, and audit logs.

### Changed

- Added feature-flag specifications and design documentation.
- Added a Docker image build target that tags images with the commit hash.
- Switched the Docker runtime image to Alpine and updated deployment documentation.

## [2.2.0] - 2026-08-26 (41bda8d)

### Added

- Development user seeding through the Python seed utility and Make target.
- Docker Compose and multi-stage container build improvements.

### Changed

- Post creation now receives the authenticated `User` and persists the author `UserID`.
- Refactored authenticated-user context handling across API handlers and services.
- Updated the frontend styling with the new design system.

## [2.1.0] - 2026-07-28 (1419d43)

### Added

- Cursor-paginated user search endpoint.
- Navbar search, search results, and debounced suggestion dropdown.

### Changed

- Improved authentication redirects and added a session-expired message after a 401 response.

## [2.0.0] - 2026-07-27 (8182eac)

### Added

- Broker-backed notification delivery for follow, like, and comment events.
- Server-Sent Events (SSE) notification streaming.
- Comprehensive API test coverage for authentication, feeds, notifications, posts, comments, and users.

### Changed

- Migrated frontend notification streaming from WebSocket to SSE.
- Centralized API error handling and improved structured logging.

## [1.1.0] - 2026-07-14 (b573169)

### Added

- Standardized RFC 9457 Problem Details API error responses.
- Stable machine-readable error codes, request IDs, and validation error details.

### Changed

- Consolidated error mapping and clarified production stack-trace behavior.

## [1.0.0] - 2026-07-06 (cfa6e47)

### Added

- Feeds homepage with the initial personal-feed experience.

## [0.5.0] - 2026-06-17 (c3ac434)

### Added

- Frontend notification system.

## [0.4.0] - 2026-05-27 (5d01f41)

### Added

- Comments, reactions, and post ownership controls.

## [0.3.0] - 2026-05-15 (104c47c)

### Added

- Social graph features and user profiles.

## [0.2.0] - 2026-05-08 (72d06ae)

### Added

- Post management and user post listings.

## [0.1.0] - 2026-04-30 (f602a2f)

### Added

- Full-stack authentication, including registration, login, logout, and session handling.

[Unreleased]: https://github.com/billykore/project-one/compare/v5.2.0...HEAD
[5.2.0]: https://github.com/billykore/project-one/compare/v5.1.0...v5.2.0
[5.1.0]: https://github.com/billykore/project-one/compare/v5.0.0...v5.1.0
[5.0.0]: https://github.com/billykore/project-one/compare/v4.1.0...v5.0.0
[4.1.0]: https://github.com/billykore/project-one/compare/v4.0.0...v4.1.0
[4.0.0]: https://github.com/billykore/project-one/compare/v3.2.0...v4.0.0
[3.2.0]: https://github.com/billykore/project-one/compare/v3.1.0...v3.2.0
[3.1.0]: https://github.com/billykore/project-one/compare/v3.0.0...v3.1.0
[3.0.0]: https://github.com/billykore/project-one/compare/v2.2.0...v3.0.0
[2.2.0]: https://github.com/billykore/project-one/compare/v2.1.0...v2.2.0
[2.1.0]: https://github.com/billykore/project-one/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/billykore/project-one/compare/v1.1.0...v2.0.0
[1.1.0]: https://github.com/billykore/project-one/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/billykore/project-one/compare/v0.5.0...v1.0.0
[0.5.0]: https://github.com/billykore/project-one/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/billykore/project-one/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/billykore/project-one/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/billykore/project-one/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/billykore/project-one/releases/tag/v0.1.0
