# Go Backend & Next.js Fullstack Template

A robust, production-ready fullstack application featuring a **Go** backend structured with **Clean Architecture** and a modern **Next.js** frontend.

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

*   **Core/Domain**: Pure Go business entities and sentinel errors. Zero dependencies on external libraries or frameworks.
*   **Core/Ports**: Dependency inversion interfaces defining how the domain interacts with the outside world.
*   **Core/UseCase**: Orchestrates business logic by implementing ports and domain models.
*   **Infrastructure (Adapters)**: Implementations of ports (GORM repositories, JWT service, Bcrypt hasher, etc.).
*   **API**: Echo handlers and DTOs (Data Transfer Objects) that interface with the UseCases.

The frontend uses the **Next.js App Router** with a clear separation of concerns in the `web/app` directory.

---

## 🛠️ Getting Started

### Prerequisites
*   Go 1.26+
*   PostgreSQL
*   Node.js & npm (for frontend)

### Backend Setup
1.  **Configure**: Copy `configs/config.yaml.example` to `configs/config.yaml` and update your database credentials.
2.  **Migrate**: Run migrations to set up the database schema.
    ```bash
    make migrate-up dsn="postgres://user:pass@host:port/db?sslmode=disable"
    ```
3.  **Run**: Start the API server on `:8080`.
    ```bash
    make run
    ```

### Frontend Setup
1.  Navigate to the `web/` directory.
2.  Install dependencies: `npm install`
3.  Run development server: `npm run dev`

### Developer Commands
| Command | Description |
| :--- | :--- |
| `make test` | Run all unit tests |
| `make mock` | Regenerate GoMock interfaces |
| `make lint` | Run golangci-lint |
| `make docs` | Regenerate Swagger documentation |
| `make check` | Run docs, vet, lint, and test in one go |

---

## 📂 Project Structure

```text
├── api/swagger/          # Auto-generated Swagger documentation
├── cmd/main.go           # Application entry point
├── configs/              # Configuration files
├── db/migrations/        # SQL migration files
├── internal/
│   ├── api/              # Handlers, DTOs, and Middlewares
│   ├── core/
│   │   ├── domain/       # Business entities
│   │   ├── ports/        # Interface definitions
│   │   └── usecase/      # Business logic implementation
│   └── infrastructure/   # Database and Service implementations
├── scripts/              # Helper shell scripts
└── web/                  # Next.js frontend application
```
