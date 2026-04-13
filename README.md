# Go DDD Backend Template (Dependency-Free)

A clean, ready-to-go Go backend template structured around **Domain-Driven Design (DDD)** and **Clean Architecture** principles. This template is designed to be **dependency-free**, utilizing only the Go standard library.

## Tech Stack
* **Language**: Go 1.25+
* **Routing & HTTP**: Standard `net/http` package
* **Logging**: Standard `log/slog` package
* **Testing**: Standard `testing` package

## Architecture
The repository adheres to strict dependency rules flowing inwards:

* **Core/Domain (`/internal/app/greeting/core/domain`)**: Pure Go business entities and sentinel errors. Zero infrastructure imports.
* **Core/Ports (`/internal/app/greeting/core/ports`)**: Dependency inversion interfaces.
* **Core/Service (`/internal/app/greeting/core/service`)**: Use-cases orchestrating ports and domain logic.
* **Adapters (`/internal/app/greeting/adapters`)**: Infrastructure implementations (HTTP Handlers using standard `net/http` with specific DTOs, and memory repositories).
* **Shared Infrastructure (`/pkg`)**: Tooling like reusable Loggers using `slog`.

## Features
* **Zero External Dependencies**: Fast builds and minimal maintenance.
* **Clean Architecture**: Strict separation of concerns (DTO mapping isolated at Handler level).
* **Domain-Driven Design**: Business logic isolated in the core domain layer.
* **Graceful Shutdown**: Built-in graceful server shutdown using standard library signals and context.
* **Preconfigured Makefile**: Simple commands for building (`make build`), running (`make run`), and testing (`make test`).

## Getting Started
1. Run `make test` to verify the project's integrity.
2. Run `make run` to start the server on `:8080`.
3. Send a GET request to `http://localhost:8080/greeting` to see the example feature in action:
   ```bash
   curl http://localhost:8080/greeting
   ```

For explicit instructions on architectural guidelines and contributing new features, refer to [AI_CONTEXT.md](./AI_CONTEXT.md).
