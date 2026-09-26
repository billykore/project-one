# Backend configuration

The backend loads `config.yaml` from a configuration directory. By default, `make run` passes `./configs`; use `make run config=/path/to/config-dir` to select another directory.

Start from the checked-in template:

```bash
cp configs/config.yaml.example configs/config.yaml
```

`config.yaml` and the `configs/keys` directory are ignored by Git because they can contain credentials and private keys.

## Settings

| YAML key | Environment variable | Purpose |
| :--- | :--- | :--- |
| `app.port` | `APP_PORT` | HTTP listen port; defaults to `8080` |
| `app.env` | `APP_ENV` | Runtime mode; `production` disables Swagger, while `debug` adds a `stack_trace` field to server-side error log entries. The RFC 9457 response body remains sanitized. |
| `app.error_type_base_url` | `APP_ERROR_TYPE_BASE_URL` | Base URL used for RFC 9457 problem type URIs |
| `database.host` | `DATABASE_HOST` | PostgreSQL host |
| `database.port` | `DATABASE_PORT` | PostgreSQL port |
| `database.user` | `DATABASE_USER` | PostgreSQL user |
| `database.password` | `DATABASE_PASSWORD` | PostgreSQL password |
| `database.dbname` | `DATABASE_DBNAME` | PostgreSQL database |
| `database.sslmode` | `DATABASE_SSLMODE` | PostgreSQL SSL mode |
| `database.max_idle_conns` | — | Maximum idle connections |
| `database.max_open_conns` | — | Maximum open connections |
| `database.conn_max_lifetime` | — | Maximum connection lifetime as a Go duration |
| `jwt.private_key_path` | `JWT_PRIVATE_KEY_PATH` | RSA private key PEM path |
| `jwt.public_key_path` | `JWT_PUBLIC_KEY_PATH` | RSA public key PEM path |
| `jwt.expiration_time` | `JWT_EXPIRATION_TIME` | Access-token lifetime as a Go duration |
| `message_broker.type` | `MESSAGE_BROKER_TYPE` | Broker identifier; use `rabbitmq` for the current application entry point |
| `message_broker.rabbitmq.url` | `MESSAGE_BROKER_RABBITMQ_URL` | AMQP connection URL |
| `message_broker.rabbitmq.exchange` | `MESSAGE_BROKER_RABBITMQ_EXCHANGE` | Notification exchange name |
| `message_broker.rabbitmq.queue` | `MESSAGE_BROKER_RABBITMQ_QUEUE` | Notification queue name |
| `feature_flags.environment` | `FEATURE_FLAGS_ENVIRONMENT` | Flag-evaluation environment: `local`, `test`, `staging`, or `production` |
| `feature_flags.refresh_interval` | `FEATURE_FLAGS_REFRESH_INTERVAL` | Interval for refreshing flag state; defaults to `30s` |
| `feature_flags.operators` | `FEATURE_FLAGS_OPERATORS` | Usernames permitted to manage feature flags |
| `monitoring.username` | `MONITORING_USERNAME` | Dedicated username for the Prometheus `/metrics` scrape |
| `monitoring.password_file` | `MONITORING_PASSWORD_FILE` | Path to the non-empty file holding the monitoring password |

Environment variable names are derived by replacing YAML dots with underscores and converting the result to uppercase. Environment values override values from the YAML file.

The config schema also contains Kafka fields for the Kafka pub/sub adapter. The current `cmd` application wiring constructs RabbitMQ publisher and subscriber implementations, so a running RabbitMQ instance and its connection settings are required.

The monitoring settings are optional as a pair. Leave both unset to make
`/metrics` reject every scrape. When enabled, set both and store the password
outside the repository; the checked-in example shows a local Compose-friendly
file path. See [deployments/README.md](../deployments/README.md#monitoring-secret)
for the Compose setup.

## JWT keys

Generate an RSA key pair for local development:

```bash
mkdir -p configs/keys
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out configs/keys/jwt-private.pem
openssl pkey -in configs/keys/jwt-private.pem -pubout -out configs/keys/jwt-public.pem
chmod 600 configs/keys/jwt-private.pem
```

The loader accepts PKCS#1 or PKCS#8 private keys and expects the public key in PKIX PEM format. Use a secret manager or protected runtime mount instead of repository files in production.

## Validation

Startup fails when the JWT key paths are empty, when the database host/user/name is missing, or when the broker type is unsupported. Invalid YAML and unreadable or malformed key files also stop startup with an error.
