# Application entry point

`cmd/main.go` is the composition root for the Project One API. It loads the
runtime configuration and RSA keys, wires the bounded contexts to PostgreSQL
and RabbitMQ adapters, starts the Echo server and notification consumer, and
performs graceful shutdown on `SIGINT` or `SIGTERM`.

Run it through the Make target:

```bash
make run
```

By default, `make run` reads `configs/config.yaml`. Point it at another
configuration directory with `make run config=/path/to/config-dir`.

Keep business rules, persistence, and HTTP handler logic in `internal/`.
Changes in this directory should normally be limited to application assembly,
startup/shutdown behavior, command-line flags, and key loading. Unit tests for
that wiring and the RSA loader live alongside the entry point.
