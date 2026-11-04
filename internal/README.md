# Internal modules

`internal` contains private application code. Business code is organised by bounded context, not by a shared technical layer.

- `identity`: users, credentials, sessions, and authentication.
- `publishing`: posts, comments, and likes.
- `social`: follows and feeds.
- `notifications`: persisted notifications and their broker event contract.
- `featureflags`: release controls.
- `operations`: health and metrics contracts.
- `pkg`: the restricted shared kernel: generic pagination, problem vocabulary, and technical ports only.

Each business context owns its `domain`, `ports`, `usecase`, `adapters`, `api`, and (where applicable) `config` packages. `pkg` owns only cross-cutting infrastructure, bootstrap configuration, and reusable HTTP mechanics. Do not add business entities, use cases, or repositories to `pkg`; the `depguard` rule in `.golangci.yml` rejects any `pkg` import of a bounded context.

Context-owned sentinel errors live in that context's `domain` package, and the context registers their problem-details mappings in `api/errors.go` from its `RegisterRoutes`. `pkg/problem` keeps only errors that cross context boundaries.

Cross-context calls should be intentional and narrow: use the upstream context's published types or a local port with only the operation required. Contexts that need a user define their own small projection in `domain` (for example publishing's `Author`) and fill it in a lookup adapter; only those adapters may import `identity/domain`, which a `depguard` rule enforces. The notification event payload is the current explicit integration contract. Shared GoMock test doubles are generated into `testkit/mocks` with `make mocks`.
