# Go Backend Template

A robust Go backend service structured around **Clean Architecture** principles.

## Tech Stack
* **Language**: Go 1.26+
* **Framework**: [Echo](https://echo.labstack.com/) (HTTP)
* **ORM**: [GORM](https://gorm.io/) (PostgreSQL)
* **Validation**: [Validator v10](https://github.com/go-playground/validator)
* **Logging**: [Zerolog](https://github.com/rs/zerolog)
* **Security**: [Bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt), [JWT](https://github.com/golang-jwt/jwt)
* **Testing**: [GoMock](https://github.com/uber/mock), [Testify](https://github.com/stretchr/testify)
* **Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate)
* **API Documentation**: [Swaggo](https://github.com/swaggo/swag)

## Architecture
The project adheres to strict dependency rules flowing inwards:

* **Core/Domain**: Pure Go business entities and sentinel errors. Zero infrastructure imports.
* **Core/Ports**: Dependency inversion interfaces.
* **Core/Service**: Use-cases orchestrating ports and domain logic.
* **Adapters**: Infrastructure implementations (Echo handlers, GORM repositories, JWT token service, etc.).
* **Pkg**: Reusable internal packages, such as the `server` package which implements the **Functional Options** pattern for highly configurable server instantiation.

## Features
* **Clean Architecture**: Strict separation of concerns.
* **Domain-Driven Design**: Business logic isolated in the core domain layer.
* **Unified Structured Logging**: Consistent use of Zerolog across all application layers, including the entrypoint.
* **User Management**: Authentication (login/logout), registration, and profile data secured with JWT and Bcrypt. Authentication tokens are managed via HttpOnly cookies.
* **Token Revocation**: Statefully track active session tokens in PostgreSQL to allow instant revocation of all user sessions upon logout.
* **Post Management**: Users can create posts with tags utilizing PostgreSQL array types.
* **Idiomatic Go Patterns**: Extensively uses idiomatic Go patterns such as Functional Options, Table-Driven tests, Sentinel Errors, and Subtests.
* **API Documentation**: Auto-generated Swagger documentation for endpoints.
* **Unit Testing**: Comprehensive unit tests for service layers, configurations, and packages using GoMock and Testify.

## Getting Started

### Database Migrations
1. Create a migration:
   ```bash
   make migrate-create name=create_users_table
   ```
2. Run migrations up:
   ```bash
   make migrate-up dsn="postgres://user:pass@host:port/db?sslmode=disable"
   ```
3. Run migrations down:
   ```bash
   make migrate-down dsn="postgres://user:pass@host:port/db?sslmode=disable"
   ```

### API Documentation
Generate Swagger documentation:
```bash
make docs
```

### Running the Application
1. Start the server:
   ```bash
   make run
   ```
2. The server will start on `:8080` (or the port specified in your configuration).

### Running Tests
1. Run all unit tests, data races, and checks:
   ```bash
   make test
   make lint
   make vet
   ```

For explicit instructions on architectural guidelines and contributing new features, refer to [AGENT.md](./AGENT.md).
