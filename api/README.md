# API contracts

`api/swagger/` contains the generated OpenAPI 2.0 contract for the Go API:

- `swagger.yaml` is the human-readable source for tooling and review.
- `swagger.json` is the JSON equivalent.
- `docs.go` registers the generated definition with the application.

The Swagger UI is available at `http://localhost:8080/swagger/index.html` when
the backend is not running with `APP_ENV=production`.

## Regenerating the contract

Handler annotations are the source of truth. After changing an annotated route,
request type, response type, or security requirement, regenerate the checked-in
artifacts from the repository root:

```bash
make docs
```

This requires the `swag` CLI. The `/metrics` scrape endpoint is deliberately
excluded from the public OpenAPI contract because it is an operational endpoint
authenticated with a dedicated monitoring credential.

The generated Postman collection in `test/api/postman-collection.json` is based
on this contract and `test/api/test-cases.md`; see [scripts/README.md](../scripts/README.md)
for regeneration details.
