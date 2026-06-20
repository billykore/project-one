# RSA JWT Key Loading Design

## Objective

Upgrade JWT handling from a shared secret to RSA signing while keeping the existing auth flow intact. The server will load the RSA key pair once during startup, pass the parsed keys into the token service, and continue exposing the same token-service behavior to the rest of the application.

This change also adds RSA key paths to config so local, test, and deployment environments can point to the correct PEM files.

## Scope

### In scope

- Replace JWT signing and verification with RSA.
- Add private/public key paths to JWT config.
- Load and parse the PEM keys once at startup.
- Pass parsed keys into the token service constructor.
- Fail fast if keys are missing, unreadable, or invalid.
- Update config examples and related docs to match the new fields.
- Add tests for config loading and JWT sign/verify behavior.

### Out of scope

- Changing the `ports.TokenService` interface.
- Adding refresh-token rotation or a token versioning scheme.
- Introducing a key management service or remote secret store.
- Changing the login, logout, websocket, or middleware call sites beyond wiring updates.

## Current State

The current implementation uses a symmetric JWT secret:

- `internal/config/config.go` exposes `jwt.secret_key` and `jwt.expiration_time`.
- `cmd/main.go` constructs the token service with the secret string.
- `internal/adapters/token/jwt_token_service.go` signs and verifies with `HS256`.

The rest of the application treats JWTs as opaque strings, so the RSA migration can stay local to config, startup wiring, and the token adapter.

## Design

### Startup key loading

`cmd/main.go` becomes the place where RSA keys are loaded and parsed.

Flow:

1. Load config as usual.
2. Read the private key PEM file from `jwt.private_key_path`.
3. Read the public key PEM file from `jwt.public_key_path`.
4. Parse the private key into an `*rsa.PrivateKey`.
5. Parse the public key into an `*rsa.PublicKey`.
6. Construct the token service with the parsed keys and access-token expiration.

This keeps file I/O and PEM parsing out of the request path and ensures any key problems stop the server before it starts serving traffic.

### Config changes

`internal/config.JWTConfig` will replace the shared secret field with two file-path fields:

- `private_key_path`
- `public_key_path`

Environment bindings should mirror those fields so the server can be configured either from YAML or environment variables.

Validation should require both paths to be present. The config layer should not parse the PEM content itself; it only checks that the paths are provided and leaves file access to startup wiring.

### Token service behavior

The token adapter should still expose the same `ports.TokenService` behavior, but internally it will work with RSA keys instead of a byte slice secret.

- `GenerateTokens` signs access tokens with the private key using `RS256`.
- `ValidateToken` verifies signatures with the public key.
- Claims, expiration behavior, and token shape remain unchanged.

The adapter should reject tokens signed with any non-RSA method.

### Key material policy

The private key must not be treated like ordinary config.

- The config file stores only a path.
- The PEM file is read from disk at startup.
- Production should source the private key from a mounted secret or equivalent secure file location.
- The public key can be distributed more broadly because it is only used for verification.

## Components

### `internal/config/config.go`

Responsibilities:

- Add `JWT.PrivateKeyPath` and `JWT.PublicKeyPath`.
- Bind `JWT_PRIVATE_KEY_PATH` and `JWT_PUBLIC_KEY_PATH`.
- Validate that both values are present.

### `configs/config.yaml.example`

Responsibilities:

- Replace the old shared-secret example with the RSA key-path fields.
- Show the expected PEM file locations for local development.

### `cmd/main.go`

Responsibilities:

- Read and parse the PEM files once at startup.
- Build the RSA-backed token service.
- Exit early if any key-related step fails.

### `internal/adapters/token/jwt_token_service.go`

Responsibilities:

- Sign access tokens with RSA.
- Verify access tokens with RSA.
- Preserve the existing token-service interface.

### Tests

Responsibilities:

- Cover config path loading and validation.
- Cover RSA token generation and validation.
- Cover startup failures for missing or malformed key files.

## Data Flow

```text
Server startup
  -> load config
  -> read private key PEM from config path
  -> read public key PEM from config path
  -> parse RSA keys
  -> construct token service

Login request
  -> use case asks token service for a token
  -> token service signs claims with RSA private key
  -> handler stores the token cookie as before

Authenticated request
  -> middleware extracts token as before
  -> token service verifies signature with RSA public key
  -> username is extracted from claims
```

## Error Handling

- Missing key-path config should fail validation before startup continues.
- Missing key files should stop the server during startup.
- Invalid PEM syntax should surface as a startup error.
- A private key/public key mismatch should fail token verification tests and should be treated as a deployment error.
- A JWT signed with the wrong algorithm should be rejected as unauthorized.

## Edge Cases

- The private key file exists but is unreadable because of permissions.
- The PEM file parses but the key type is not RSA.
- The public key file does not match the private key used for signing.
- Existing tokens become invalid after cutover because the signing algorithm changes.
- Local development needs a simple, documented default path for generated PEM files.

## Testing Strategy

### Config tests

- Loading a config file with both RSA key paths should succeed.
- Environment variables should override file values for both paths.
- Missing private key path should fail validation.
- Missing public key path should fail validation.

### Token service tests

- Tokens signed with the private key should validate with the public key.
- Tokens signed with a different key pair should fail verification.
- Tokens signed with a non-RSA method should be rejected.

### Startup tests

- Main startup wiring should fail if the private key file cannot be read.
- Main startup wiring should fail if the public key file cannot be parsed.

## Implementation Notes

- Keep the token-service interface unchanged unless a test proves a hard need for an API change.
- Prefer loading parsed keys once instead of re-reading files on every token operation.
- Keep the key-loading code close to startup wiring so the failure mode is obvious.
- Update documentation that still mentions a JWT secret so the repo does not describe two different auth models at once.

## Success Criteria

- JWTs are signed with RSA instead of a shared secret.
- The server loads RSA keys once at startup and fails fast on bad key material.
- JWT config uses key paths instead of a secret string.
- The existing auth flow continues to work without call-site churn.
- Tests cover config, startup failures, and RSA token verification.