# AI Agent Development Guidelines

This project is a **Go backend service template** following Domain-Driven Design (DDD) and Clean Architecture principles. This template is designed to be **dependency-free**, utilizing only the Go standard library for its core functionality and HTTP handling.

## Architecture Overview

```
cmd/greeting/main.go          ← Entrypoint: wires all dependencies
pkg/logger/                     ← Shared infrastructure (std library log/slog)
internal/app/greeting/
├── core/                       ← THE CORE — pure business logic, zero infra imports
│   ├── domain/                 ← Entities, value objects, domain errors
│   ├── ports/                  ← Interfaces (driven + driving ports)
│   └── service/                ← Application services implementing driving ports
└── adapters/                   ← INFRASTRUCTURE — implements driven ports
    ├── handler/                ← HTTP handlers (driving adapter) + DTOs
    └── repository/             ← Data persistence (driven adapter)
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

- **Entities** are plain Go structs with **no framework tags** (no `json:`, `gorm:`, `db:` tags).
- Entities contain **domain validation** methods (e.g., `func (g *Greeting) Validate() error`).
- **Sentinel errors** (e.g., `ErrNotFound`, `ErrInvalidInput`) are defined here for use across layers.
- Domain types are the shared language — all layers reference them.

## Ports Layer (`core/ports`)

- **Driving ports** (e.g., `GreetingService`) define what the application *can do* — implemented by services.
- **Driven ports** (e.g., `GreetingRepository`) define what the application *needs* — implemented by adapters/infrastructure.
- Keep interfaces small and focused (Interface Segregation Principle).

## Service Layer (`core/service`)

- Implements driving ports.
- Orchestrates domain logic by calling driven ports.
- **Must never import infrastructure packages** (like `net/http` for handler logic).
- All infrastructure is injected via constructor using port interfaces.
- Returns domain errors for business rule violations.

## Adapter Layer (`adapters/`)

### Handlers (`adapters/handler`)
- HTTP handlers are the **driving adapters** — they translate HTTP into service calls.
- Uses the standard library **`net/http`** package.
- Handler methods typically use `http.ResponseWriter` and `*http.Request`.
- Use `json.NewDecoder()` for request deserialization and `json.NewEncoder()` for responses.
- Use **DTOs** (`dto.go`) for request/response serialization — never decode directly into domain entities.
- Map domain errors to appropriate HTTP status codes using `errors.Is()`.

### Repositories (`adapters/repository`)
- Repositories are **driven adapters** — they implement persistence port interfaces.
- Return **domain sentinel errors**, not raw strings.
- Repository implementations may use framework-specific tags on internal model structs (not domain entities).

## Error Handling Conventions

1. **Domain errors** — Defined as sentinel errors in `core/domain/errors.go`.
2. **Service layer** — Wraps infrastructure errors with `fmt.Errorf("context: %w", err)` and returns domain errors for business violations.
3. **Handler layer** — Uses `errors.Is()` to match domain errors and map to HTTP status codes.
4. **Never expose internal error details** to the client — use generic messages for unexpected errors.

## Adding a New Feature (Step-by-Step)

1. **Define the domain entity** in `core/domain/` (plain struct, no tags, with `Validate()` method).
2. **Add sentinel errors** to `core/domain/errors.go` if needed.
3. **Define port interfaces** in `core/ports/` (repository interface + service interface).
4. **Implement the service** in `core/service/` using only ports for dependencies.
5. **Write unit tests** in `core/service/` using manual mock implementations of ports.
6. **Create DTOs** in `adapters/handler/dto.go` with mapping functions to/from domain.
7. **Implement the HTTP handler** in `adapters/handler/` using standard `net/http`.
8. **Implement the repository** in `adapters/repository/`.
9. **Wire everything** in `cmd/greeting/main.go`.

## Testing Strategy

- **Unit tests** live alongside the code they test (e.g., `greeting_service_test.go`).
- **Manual mocks**: Since this is a dependency-free template, use manual mock implementations of interfaces in your tests.
- Test business logic in isolation from infrastructure.
- Use `go test -v ./...` to run all tests (note: directories starting with `_` like `greeting` may need explicit paths for `go test`).

## Coding Standards

- **Idiomatic Go**: return early on errors, keep functions short, use named return values sparingly.
- **Exported types**: handler structs, constructors, and port interfaces should be exported.
- **Context propagation**: always pass `context.Context` as the first argument.
- **Configuration**: read from environment variables or flags in `main.go`, pass values via constructors.
- **No global mutable state**: avoid package-level `var` for configuration. Inject everything.
