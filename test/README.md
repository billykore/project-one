# Test assets and reports

`test/` contains artifacts that supplement the Go tests under `internal/` and
the frontend tests under `web/tests/`.

- `api/test-cases.md` is the human-maintained API test matrix.
- `api/postman-collection.json` is the generated Postman v2.1 collection.
- `coverage/` holds local backend coverage output from `make test-cover`.
- `reports/` contains recorded manual and integration test reports.
- `screenshots/` contains evidence captured for documented UI flows.

Regenerate the Postman collection after changing the test matrix or Swagger
contract:

```bash
python3 scripts/generate-postman-collection.py
```

Generate a local backend coverage report with:

```bash
make test-cover
```

The resulting HTML file is `test/coverage/coverage.html`. Coverage output and
other machine-local artifacts should not be treated as source-of-truth test
evidence unless intentionally added to a change.

