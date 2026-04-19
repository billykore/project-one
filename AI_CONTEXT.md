# AI Agent Development Guidelines

This project is a **Go backend service** following Domain-Driven Design (DDD) and Clean Architecture principles. It utilizes a modern tech stack including Echo for HTTP handling and GORM for database persistence.

## Architecture Overview

```
cmd/user/main.go              ← Entrypoint: wires all dependencies
pkg/logger/                   ← Shared infrastructure (zerolog)
internal/app/user/
├── core/                     ← THE CORE — pure business logic, zero infra imports
│   ├── domain/               ← Entities, value objects, domain errors
│   ├── ports/                ← Interfaces (driven + driving ports)
│   └── service/              ← Application services implementing driving ports
└── adapters/                 ← INFRASTRUCTURE — implements driven ports
    ├── handler/              ← Echo handlers (driving adapter) + DTOs
    ├── repository/           ← GORM repositories (driven adapter)
    ├── hasher/               ← Bcrypt hasher implementation
    ├── logger/               ← Zerolog logger implementation
    └── token/                ← JWT token service implementation
```

## The Dependency Rule

Dependencies ALWAYS flow **inward**. Outer layers depend on inner layers, never the reverse.

```
Adapters → Ports ← Services → Domain
    ↑                            ↑
    └────── NEVER imports ───────┘
```

**Strict rules:**
1. `core/domain` imports **nothing** from this project — only the Go standard library.
2. `core/ports` imports only `core/domain`.
3. `core/service` imports only `core/domain` and `core/ports`.
4. `adapters/` imports `core/domain` and `core/ports` — **never** `core/service`.
5. `pkg/` is shared infrastructure — may be imported by `adapters/` and `cmd/`, never by `core/`.

## Domain Layer (`core/domain`)

- **Entities** are Go structs with minimal to no framework tags. Domain-level validation is performed here.
- **Sentinel errors** (e.g., `ErrUserNotFound`) are defined here for use across layers.
- Domain types are the shared language — all layers reference them.

## Ports Layer (`core/ports`)

- **Driving ports** (e.g., `LoginService`) define what the application *can do* — implemented by services.
- **Driven ports** (e.g., `UserRepository`, `Hasher`, `TokenService`) define what the application *needs* — implemented by adapters/infrastructure.

## Service Layer (`core/service`)

- Implements driving ports and orchestrates domain logic by calling driven ports.
- **Must never import infrastructure packages** (like `echo` or `gorm`).
- All infrastructure is injected via constructor using port interfaces.
- Returns domain errors for business rule violations.

## Adapter Layer (`adapters/`)

### Handlers (`adapters/handler`)
- HTTP handlers are the **driving adapters** using the **Echo** framework.
- Use **DTOs** (`dto.go`) for request/response serialization — never decode directly into domain entities.
- Map domain errors to appropriate HTTP status codes (e.g., `domain.ErrInvalidCredentials` to `401 Unauthorized`).

### Repositories (`adapters/repository`)
- Repositories are **driven adapters** using **GORM** for persistence.
- Map GORM-specific errors (like `gorm.ErrRecordNotFound`) to domain sentinel errors.

## Error Handling Conventions

1. **Domain errors** — Defined as sentinel errors in `core/domain/errors.go`.
2. **Service layer** — Returns domain errors directly or wraps unexpected errors.
3. **Handler layer** — Uses `errors.Is()` to match domain errors and map to HTTP status codes.

## Adding a New Feature (Step-by-Step)

1. **Define the domain entity** in `core/domain/`.
2. **Add sentinel errors** to `core/domain/errors.go` if needed.
3. **Define port interfaces** in `core/ports/`.
4. **Implement the service** in `core/service/`.
5. **Generate mocks** for new ports using `mockgen`.
6. **Write unit tests** for the service in `core/service/` using generated mocks.
7. **Create DTOs** in `adapters/handler/dto/`.
8. **Implement the HTTP handler** in `adapters/handler/` using Echo.
9. **Implement the repository** in `adapters/repository/` using GORM.
10. **Wire everything** in `cmd/<app>/main.go`.

## Testing Strategy

- **Unit tests** live alongside the code they test.
- **Mocks**: Use `mockgen` (GoMock) to generate mocks for port interfaces.
- Only test **domain** and **service** layers in unit tests. Service tests must mock all port dependencies.

## Coding Standards

- **Idiomatic Go**: return early on errors, keep functions short.
- **Context propagation**: always pass `context.Context` as the first argument.
- **Dependency Injection**: Always inject dependencies via constructors.
