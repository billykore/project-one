<!--
  Sync Impact Report
  ==================
  Version change: 0.0.0 (template) → 1.0.0
  Modified principles: N/A (initial creation — all principles are new)
  Added sections:
    - Core Principles (I–IV: Code Quality, Testing, UX Consistency, Performance)
    - Security & Authentication
    - Development Workflow & Quality Gates
  Removed sections: None
  Templates requiring updates:
    - .specify/templates/plan-template.md      ✅ aligned (Constitution Check gate is generic)
    - .specify/templates/spec-template.md       ✅ aligned (user stories, requirements, success criteria)
    - .specify/templates/tasks-template.md      ✅ aligned (test-first, phases, checkpoints)
    - README.md                                 ✅ aligned (architecture, commands, conventions)
  Follow-up TODOs: None
-->

# Project1 Constitution

## Core Principles

### I. Code Quality & Clean Architecture (NON-NEGOTIABLE)

The backend MUST follow Clean Architecture with strict layer separation:

- **Domain layer** (`internal/core/domain/`): Pure Go entities and sentinel errors. Zero dependencies
  on external libraries or frameworks (no GORM annotations, no Echo imports).
- **Ports layer** (`internal/core/ports/`): Interface definitions using dependency inversion. Every
  external interaction (database, auth, hashing) MUST be defined as an interface here.
- **UseCase layer** (`internal/core/usecase/`): Orchestrates business logic against ports only.
  Never imports adapter implementations directly.
- **Adapters layer** (`internal/adapters/`): Concrete implementations of port interfaces.
  GORM repositories, JWT services, Bcrypt hashers all live here.
- **API layer** (`internal/api/`): Echo handlers, DTOs, and middleware. Handlers delegate to
  use cases; they do NOT contain business logic.

All Go code MUST pass `make check` (docs + vet + lint + test) before merging. The `golangci-lint`
configuration is the authoritative style guide. No commented-out code or dead imports may persist.

### II. Testing Standards

Testing is mandatory for all use cases and adapters:

- **Unit tests**: Every use case MUST have unit tests using GoMock-generated mocks
  (`internal/core/usecase/mocks/`). Run `make mock` after interface changes.
- **Integration tests**: Adapter implementations (repositories, auth services) SHOULD have
  integration tests against real infrastructure (test database, etc.).
- **Frontend tests**: Components and utilities under `web/tests/` run with Vitest. New
  features SHOULD include corresponding test files.
- **Test-first discipline**: For bug fixes, write a failing test that reproduces the bug
  before applying the fix. For new features, tests SHOULD be written alongside or before
  implementation code.
- **Coverage**: Use `test/coverage/` for coverage reports. Coverage must not decrease
  with new PRs. Critical paths (auth, data mutation, payment) MUST maintain >80% coverage.

Tests MUST be runnable with a single command (`make test` for backend, `npm test` for frontend)
and MUST pass in CI.

### III. User Experience Consistency

The frontend MUST deliver a consistent, accessible user experience:

- **Design system compliance**: All UI MUST use the project's glassmorphism design system
  (defined in `web/app/globals.css`). Custom CSS outside the design tokens requires
  explicit justification.
- **Component reuse**: New UI elements MUST use existing shadcn/ui + Radix primitives
  before introducing new dependencies. Check `web/components/` for existing patterns.
- **Responsive design**: All pages MUST be responsive (mobile-first). Breakpoints follow
  Tailwind CSS 4 defaults.
- **Accessibility**: Interactive elements MUST have focus states, ARIA labels where
  appropriate, and keyboard navigation support. Color contrast MUST meet WCAG AA minimum.
- **Loading & error states**: Every data-fetching UI MUST handle loading, empty, error,
  and success states. Skeleton loaders (`web/components/ui/skeleton.tsx`) are the
  standard loading pattern.
- **Guest UX**: Unauthenticated users MUST see appropriate guards (disabled interactions
  with tooltips, not broken pages or confusing errors).

### IV. Performance Requirements

The application MUST meet these performance standards:

- **API response time**: p95 latency MUST be <200ms for read endpoints and <500ms for
  write endpoints under normal load.
- **Database queries**: Use GORM preloading and indexing strategies. Every new query
  MUST be reviewed for N+1 problems. Migrations MUST include appropriate indexes
  (see `db/migrations/` for index naming conventions).
- **Frontend**: Next.js pages SHOULD leverage Server Components by default; `"use client"`
  boundaries MUST be justified. Bundle size for any page SHOULD stay under 200KB
  (uncompressed JS).
- **WebSocket efficiency**: The notification WebSocket (`GET /ws`) MUST only push
  new notifications for the authenticated user, not broadcast globally.
- **Caching & pagination**: List endpoints MUST support cursor-based or offset
  pagination. No unbounded queries — every `SELECT` without `LIMIT` is a bug.

## Security & Authentication

- **Password storage**: Passwords MUST be hashed with Bcrypt before storage. Clear-text
  passwords MUST never appear in logs, error messages, or API responses.
- **Token management**: JWTs MUST have reasonable expiration times. Access tokens use
  short-lived expiry; refresh token rotation SHOULD be implemented.
- **Input validation**: All API inputs MUST be validated server-side using `validator/v10`
  struct tags. Never trust client-side validation alone.
- **Authorization**: Every protected endpoint MUST verify the JWT and check ownership
  before mutating resources (e.g., only the post author may delete their post).
- **HTTPS**: Production deployments MUST use HTTPS. Sensitive endpoints (login, register,
  change-password) MUST reject non-HTTPS requests in production.

## Development Workflow & Quality Gates

- **Commit convention**: All commits MUST follow the format:
  `<type>(<scope>): <description>` where type is one of `feat`, `fix`, `chore`,
  `refactor`, `docs`, `test`. Enforced by the `prepare-commit-msg` Git hook
  (activate with `make githooks`).
- **Code review**: Every PR MUST be reviewed by at least one other developer before
  merge. Reviewers MUST verify constitution compliance.
- **Pre-merge gates**: `make check` and `npm test` MUST pass. No warnings allowed.
- **Branch naming**: Feature branches SHOULD use the pattern `###-short-description`
  or reference a ticket ID.
- **Documentation**: New API endpoints MUST include Swagger annotations
  (`api/swagger/`). Run `make docs` to regenerate.

## Governance

This constitution supersedes all other development practices and conventions. Any
deviation from its principles MUST be documented in the implementation plan's
"Complexity Tracking" section with a clear justification and analysis of simpler
alternatives that were rejected.

**Amendment procedure**: Proposed changes MUST be submitted as a PR modifying this
file. Amendments require review and approval from the project maintainer. The
version number MUST be incremented per semantic versioning rules (MAJOR for
principle removals/redefinitions, MINOR for new principles, PATCH for clarifications).

**Compliance**: All PRs and code reviews MUST verify compliance with these principles.
The Constitution Check section of every implementation plan MUST address each
applicable principle. Complexity or violations MUST be justified explicitly.

**Version**: 1.0.0 | **Ratified**: 2026-07-07 | **Last Amended**: 2026-07-07
