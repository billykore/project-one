# Build artifacts and packaging

`build/` is reserved for build output and packaging/CI support files.

- `build/bin/main` is the backend executable produced by `make build`.
- `build/package/` is reserved for package-specific build assets.
- `build/ci/` is reserved for CI support assets when a workflow cannot live in
  its provider-required location.

The executable is generated locally and must not be committed. Rebuild it with:

```bash
make build
```

Remove generated backend artifacts with `make clean`. Container build inputs
live at the repository root in `Dockerfile`; the local multi-service runtime is
defined in [deployments/compose.yml](../deployments/compose.yml).
