# Development and verification scripts

These scripts support contract generation and performance/operational checks.
Run them from the repository root.

| Script | Purpose | Typical invocation |
| :--- | :--- | :--- |
| `generate-postman-collection.py` | Generates the Postman v2.1 collection from the API test-case table and Swagger contract | `python3 scripts/generate-postman-collection.py` |
| `verify-request-latency.py` | Logs in, measures authenticated notification reads and post writes, and enforces p95 budgets | `PROJECT1_EMAIL=... PROJECT1_PASSWORD=... python3 scripts/verify-request-latency.py` |
| `verify-observability-latency.py` | Verifies `/healthz`, `/status`, and authenticated `/metrics` remain within the operational latency budget | `python3 scripts/verify-observability-latency.py` |
| `benchmark-post-cqrs.sh` | Measures representative post read and like-write p95 latency with `curl` | `BASE_URL=... PUBLIC_POST_ID=... AUTHOR_POST_ID=... ACCESS_TOKEN=... bash scripts/benchmark-post-cqrs.sh` |

The Postman generator requires PyYAML. Install it with `pip install pyyaml` if
it is not already available. It writes to `test/api/postman-collection.json` by
default; use its `--input-md`, `--swagger`, or `--output` options to override
the defaults.

The verification scripts call a running API and may create test posts. Use a
development or disposable dataset, never a production account or database.
