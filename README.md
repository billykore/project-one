# Go Backend & Next.js Fullstack App

A robust, production-ready fullstack application featuring a **Go** backend structured with **Clean Architecture** and a modern **Next.js** frontend.

## ✨ Features

* **User Management**: Registration and login with Email/Username and Password (JWT-based).
* **Social Connectivity**: Follow and unfollow users to build a personal feed.
* **Content Creation**: Create and view posts with real-time feedback.
* **Notifications**: Frontend notification system with panel and dropdown UI for live updates.
* **Guest UX**: Read-only guest session guards with disabled post interaction states and tooltips.
* **Profile Experience**: User profile pages with root-level dynamic routing and dashboard links.
* **Secure Password Management**: Backend change-password API, validation, and encrypted password storage.
* **Idempotent Post Likes**: Like/unlike behavior has been refactored for consistent idempotent interactions.
* **Post Ownership Controls**: Delete post actions are restricted to the author and UI reflects authorization.
* **Clean Architecture**: Backend strictly follows separation of concerns for testability and maintainability.
* **Modern Frontend**: Server-side rendering and interactive UI using Next.js 16 and Tailwind 4.

## 🛠️ Recent Work

Recent repository work includes:

* Adding notification system components and utilities in the frontend.
* Improving guest session handling and disabled like button UX.
* Making post detail pages publicly accessible while keeping auth-aware headers.
* Building user profile routes and profile navigation flows.
* Implementing a secure change password API endpoint with backend validation.
* Updating the follow model to include `FollowerID` and `FollowedID` in domain/repository layers.
* Refactoring likes to be idempotent and restricting delete post options to the author.

## 🚀 Tech Stack

### Backend

* **Language**: Go 1.26+
* **Framework**: [Echo](https://echo.labstack.com/) (HTTP)
* **ORM**: [GORM](https://gorm.io/) with PostgreSQL
* **Validation**: [Validator v10](https://github.com/go-playground/validator)
* **Logging**: [Zerolog](https://github.com/rs/zerolog)
* **Security**: [Bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt), [JWT v5](https://github.com/golang-jwt/jwt)
* **Testing**: [GoMock](https://github.com/uber/mock), [Testify](https://github.com/stretchr/testify)
* **API Documentation**: [Swaggo](https://github.com/swaggo/swag)

### Frontend

* **Framework**: [Next.js 16.2.4](https://nextjs.org/) (App Router)
* **Library**: [React 19](https://react.dev/)
* **Styling**: [Tailwind CSS 4](https://tailwindcss.com/)
* **Language**: [TypeScript](https://www.typescriptlang.org/)
* **Components**: UI components built with Radix UI primitives.

---

## 🏗️ Architecture

The backend follows **Clean Architecture** principles to ensure separation of concerns and maintainability:

* **Core/Domain**: Pure Go business entities and sentinel errors. Zero dependencies on external libraries or frameworks.
* **Core/Ports**: Dependency inversion interfaces defining how the domain interacts with the outside world.
* **Core/UseCase**: Orchestrates business logic by implementing ports and domain models.
* **Adapters**: Concrete implementations of ports (GORM repositories, JWT service, Bcrypt hasher, etc.).
* **API**: Echo handlers, DTOs (Data Transfer Objects), and Middleware.

The frontend uses the **Next.js App Router** with layouts, pages, and components in the `web/app` directory, and API clients / utilities in `web/lib`.

---

## 🛠️ Getting Started

### Prerequisites

* **Go**: 1.26+
* **PostgreSQL**: For database storage
* **Node.js & npm**: For running the Next.js frontend
* **Additional Tooling** (Required for development commands):
  * [golang-migrate CLI](https://github.com/golang-migrate/migrate): For database migrations (`make migrate-*`)
  * [swag CLI](https://github.com/swaggo/swag): For generating Swagger API docs (`make docs`)
  * [golangci-lint](https://golangci-lint.run/): For static analysis (`make lint`)

### Backend Setup

1. **Configure**: Copy `configs/config.yaml.example` to `configs/config.yaml` and update your database credentials.
2. **Migrate**: Run migrations to set up the database schema.

    ```bash
    make migrate-up dsn="postgres://user:pass@host:port/db?sslmode=disable"
    ```

3. **Run**: Start the API server on `:8080`.

    ```bash
    make run
    ```

### Frontend Setup

1. Navigate to the `web/` directory.
2. Install dependencies: `npm install`
3. Run development server: `npm run dev`

### Developer Commands

| Command | Description |
| :--- | :--- |
| `make build` | Compile the backend application binary to `bin/main` |
| `make run` | Compile and run the backend API server |
| `make test` | Run all unit tests |
| `make mock` | Regenerate GoMock interfaces in `internal/core/usecase/mocks/` |
| `make vet` | Run `go vet` |
| `make lint` | Run static analysis via `golangci-lint` |
| `make docs` | Regenerate Swagger API documentation |
| `make migrate-create name=...` | Create a new SQL migration file |
| `make migrate-up dsn=...` | Run database migrations up |
| `make migrate-down dsn=...` | Run database migrations down |
| `make githooks` | Configure git to use local pre-commit and pre-push hooks |
| `make check` | Run docs, vet, lint, and test in one go |
| `make clean` | Remove backend build artifacts from `bin/` |

### Real-time Notifications (WebSocket)

- Endpoint: `GET /ws`
- Auth: `Authorization: Bearer <access_token>` during WebSocket handshake
- Behavior: streams only new notifications for the authenticated user
- Historical notifications: use `GET /notifications`

---

## 📂 Project Structure

```text
├── api/swagger/          # Auto-generated Swagger documentation
├── bin/                  # Directory containing compiled backend binary
├── cmd/main.go           # Application entry point
├── configs/              # Configuration files (config.yaml)
├── db/migrations/        # SQL migration files
├── deployments/          # Deployment configurations and templates
├── docs/                 # Documentation (plans, specifications, and tasks)
├── githooks/             # Local Git hooks (pre-commit, pre-push)
├── internal/
│   ├── api/              # Handlers, DTOs, and Middlewares
│   ├── core/
│   │   ├── domain/       # Business entities
│   │   ├── ports/        # Interface definitions
│   │   └── usecase/      # Business logic implementation
│   ├── adapters/         # Implementation of ports (DB, services)
│   └── config/           # Application configuration logic
├── scripts/              # Helper shell scripts for make commands
└── web/                  # Next.js frontend application
```
