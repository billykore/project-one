# Docker Compose deployment

`compose.yml` defines the complete local Project One stack. It is intended for development and integration testing rather than as a hardened production deployment.

## Services

| Service | Image/build | Published port | Purpose |
| :--- | :--- | :--- | :--- |
| `postgres` | `postgres:17-alpine` | `5432` | Application data |
| `rabbitmq` | `rabbitmq:4-alpine` | `5672` | Notification event transport |
| `keys` | `alpine:3.22` | None | Copies JWT keys into a protected shared volume |
| `backend` | Root `Dockerfile` | `8080` | Go API and notification consumer; health-checked through `/status` |
| `frontend` | `web/Dockerfile` | `3000` | Next.js standalone server |

PostgreSQL, RabbitMQ, and the backend have health checks. The backend starts only after the database and broker are healthy and JWT key preparation has completed; the frontend starts after the backend container is created. The backend probe uses Alpine's built-in `wget` to request `/status`.

## Prerequisites

- Docker with Compose support
- An RSA private/public key pair at `configs/keys/jwt-private.pem` and `configs/keys/jwt-public.pem`

Generate development keys from the repository root:

```bash
mkdir -p configs/keys
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out configs/keys/jwt-private.pem
openssl pkey -in configs/keys/jwt-private.pem -pubout -out configs/keys/jwt-public.pem
```

The key directory is ignored by Git. Do not commit private keys.

## Lifecycle

Run these commands from the repository root:

```bash
make compose-up       # build images and start in the background
make compose-stop     # stop containers, preserving them
make compose-start    # restart existing stopped containers
make compose-down     # remove containers and network, retaining volumes
```

To inspect the stack directly:

```bash
docker compose -f deployments/compose.yml ps
docker compose -f deployments/compose.yml logs -f backend frontend
```

The frontend is served at <http://localhost:3000>, and the API is served at <http://localhost:8080>. In non-production mode, Swagger UI is at <http://localhost:8080/swagger/index.html>.

## Configuration

The Compose file supplies backend configuration through environment variables. Override database and RabbitMQ credentials in the invoking shell or in a root `.env` file read by Docker Compose:

```dotenv
POSTGRES_DB=project1
POSTGRES_USER=project1
POSTGRES_PASSWORD=change-me
RABBITMQ_USER=project1
RABBITMQ_PASSWORD=change-me
```

These variables configure both the infrastructure containers and the backend connection strings. The frontend receives `API_URL=http://backend:8080`, which is reachable only within the Compose network.

## Database initialization and volumes

`init-db.sh` runs every `db/migrations/*.up.sql` file, in filename order, when PostgreSQL initializes a new data volume. PostgreSQL only runs scripts under `/docker-entrypoint-initdb.d` for an empty data directory.

For schema changes against an existing `postgres-data` volume, run the migration CLI explicitly. Replace the placeholders below with the same database name, user, and password that initialized the volume (values in a Compose `.env` file are not automatically exported to your shell):

```bash
make migrate-up dsn='postgres://<user>:<password>@localhost:5432/<database>?sslmode=disable'
```

The stack uses three named volumes:

- `postgres-data` for PostgreSQL data
- `rabbitmq-data` for broker data
- `jwt-keys` for the runtime copy of the RSA keys

`make compose-down` preserves these volumes. Running `docker compose -f deployments/compose.yml down -v` deletes all three volumes and permanently removes their local data.

## Container images

The backend image uses a Go 1.26.2 Alpine build stage and an Alpine runtime with BusyBox `wget` for its health probe; it executes as UID/GID 65532. The frontend image uses Node.js 24 Alpine, Next.js standalone output, and the unprivileged `node` user.
