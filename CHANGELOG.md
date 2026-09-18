# Changelog

All notable changes to Project One are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
