# Tasks: External Log Aggregation

**Input**: Design documents from `/specs/008-external-log-aggregation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, [deployment contract](./contracts/log-aggregation-deployment.md), quickstart.md

**Tests**: Add focused middleware tests before implementation. Preserve backend coverage, run frontend Vitest, enforce the constitution's read/write p95 limits, and run the Compose privacy and outage acceptance checks in quickstart.md.

**Organization**: Shared Loki/Alloy work is foundational; remaining tasks are grouped by user story.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Define deployment-owned Loki, collector, and Grafana assets.

- [X] T001 [P] Create the private single-process Loki configuration with local storage and retention settings in `deployments/loki.yml`.
- [X] T002 [P] Create the backend-only Grafana Alloy pipeline in `deployments/alloy/config.alloy` with Docker discovery, JSON parsing, static labels, an allowlisted event projection, secret filtering, instance metadata, and the default best-effort writer without WAL or an experimental queue.
- [X] T003 [P] Provision the non-default Loki datasource using Loki's internal base URL in `deployments/grafana/provisioning/datasources/project-one-loki.yml`.
- [X] T004 [P] Run `make test-cover` before implementation and record the `go tool cover -func=test/coverage/coverage.out | tail -1` result in `specs/008-external-log-aggregation/coverage-baseline.txt`.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Connect the existing Compose stack to Loki and Alloy without putting log delivery in the backend lifecycle.

**⚠️ CRITICAL**: No user story starts until this phase is complete.

- [X] T005 Wire Loki and Alloy into `deployments/compose.yml` with a private-only Loki service, backend-only Docker discovery, an `APP_ENV` label source, named Loki/Alloy volumes, and no backend dependency on either observability service.

**Checkpoint**: `docker compose -f deployments/compose.yml config` succeeds; Loki has no host port and no Go package depends on Loki.

---

## Phase 3: User Story 1 - Investigate Application Events (Priority: P1) 🎯 MVP

**Goal**: Operators can search safe, structured backend events in Grafana Explore by time, severity, and request correlation.

**Independent Test**: Start the stack, make normal and failing requests with known request IDs, and find their safe events in Grafana Explore within 30 seconds using `specs/008-external-log-aggregation/quickstart.md`.

### Tests for User Story 1

- [X] T006 [P] [US1] Add failing tests for safe completion-event fields, resolved route templates, request IDs, and exclusion of raw URL/query/address/user-agent fields in `internal/api/middleware/request_logging_test.go`.

### Implementation for User Story 1

- [X] T007 [US1] Implement the structured completion-logging middleware in `internal/api/middleware/request_logging.go` using only fields allowed by `specs/008-external-log-aggregation/data-model.md`.
- [X] T008 [US1] Replace Echo's default request logger with the safe completion middleware in `cmd/main.go`, preserving request IDs, recovery, metrics, and error-handler ordering.
- [X] T009 [US1] Execute a 100-event delivery check containing 20 distinct request IDs using `specs/008-external-log-aggregation/quickstart.md`; verify every event appears within 30 seconds and all 20 contexts can be located within two minutes.
- [X] T010 [US1] Run the prohibited-value sentinel matrix in `specs/008-external-log-aggregation/quickstart.md` for passwords, tokens, cookies, account identifiers, query strings, request bodies, private messages, raw errors, and stack traces; verify zero matches in Loki and Alloy diagnostics.

**Checkpoint**: Grafana finds every expected representative event; the Loki stream contains no raw request URL, query string, account identifier, or error text.

---

## Phase 4: User Story 2 - Configure a Deployment Safely (Priority: P2)

**Goal**: Maintainers can use the private local Loki service or a protected external destination without embedding secrets in application artifacts.

**Independent Test**: Validate the local Compose configuration and a protected-destination configuration using an ignored read-only secret reference; absent or rejected destination settings leave backend local logging and user workflows intact.

### Implementation for User Story 2

- [X] T011 [US2] Parameterize the Loki base URL, optional tenant identity, and optional read-only credential-file reference in `deployments/compose.yml` and `deployments/alloy/config.alloy`, preserving the private local Loki default and never exposing secret contents.
- [X] T012 [US2] Add default, deployment-disabled, and rejected protected-destination validation scenarios to `specs/008-external-log-aggregation/quickstart.md`, defining Alloy's running state as the enabled/disabled signal.
- [X] T013 [US2] Run the configuration scenarios in `specs/008-external-log-aggregation/quickstart.md`; verify delivery failures appear in `docker compose -f deployments/compose.yml logs alloy` without credentials and backend stderr remains independently available through `docker compose -f deployments/compose.yml logs backend`.

**Checkpoint**: Valid settings deliver logs; missing or invalid protected-destination settings do not stop normal backend operation or expose credentials.

---

## Phase 5: User Story 3 - Remain Available During a Destination Outage (Priority: P2)

**Goal**: A slow or unavailable aggregation system cannot become a backend outage.

**Independent Test**: Stop Loki and Alloy for 15 minutes, exercise publishing, authentication, and notification workflows, and confirm normal outcomes while backend stderr remains available.

### Implementation for User Story 3

- [X] T014 [US3] Verify `deployments/compose.yml` keeps backend startup, health checks, and shutdown independent from Loki and Alloy, with recovery behavior confined to observability services.
- [X] T015 [US3] Execute and document the 15-minute outage acceptance scenario, recovery, and best-effort lost-event boundary in `specs/008-external-log-aggregation/quickstart.md`.

**Checkpoint**: Tested workflows retain their normal outcome during the outage; later events resume after recovery without replay guarantees.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Finalize documentation and validate the full change set.

- [X] T016 [P] Update Loki, Alloy, Grafana Explore, local secret, private-network, and troubleshooting documentation in `README.md` and `deployments/README.md`.
- [X] T017 [P] Create a standard-library read/write p95 acceptance verifier in `scripts/verify-request-latency.py` that exercises representative authenticated read and write operations and fails when read p95 is at least 200 ms or write p95 is at least 500 ms.
- [X] T018 Run `gofmt` for `internal/api/middleware/request_logging.go`, then run every automated and Compose check in `specs/008-external-log-aggregation/quickstart.md`, including `make check`, `npm test -- --run`, `scripts/verify-request-latency.py`, and a final `make test-cover`; compare the final total coverage with `specs/008-external-log-aggregation/coverage-baseline.txt` and fail validation if coverage decreased.

---

## Dependencies & Execution Order

- **Setup**: T001–T004 can run in parallel; T004 must finish before any Go code changes.
- **Foundation**: T005 depends on T001–T004 and blocks every story.
- **US1**: T006 → T007 → T008; T009 depends on T005 and T008; T010 depends on T009. This is the MVP.
- **US2**: T011 depends on T005; T012–T013 depend on T011.
- **US3**: T014 depends on T005; T015 depends on T014.
- **Polish**: T016–T018 follow the selected user-story phases; T018 depends on T017.

## Parallel Opportunities

```text
T001 deployments/loki.yml
T002 deployments/alloy/config.alloy
T003 deployments/grafana/provisioning/datasources/project-one-loki.yml
T004 specs/008-external-log-aggregation/coverage-baseline.txt
```

After T005, T006, T011, and T014 can proceed in parallel. T016 and T017 can run alongside late story work after configuration and representative operations are stable.

## Implementation Strategy

### MVP First

1. Complete T001–T005 for the coverage baseline and private backend-only collection path.
2. Complete T006–T010 for safe, searchable events and privacy validation.
3. Verify a privacy sentinel cannot reach Loki before moving on.

### Incremental Delivery

1. Shared foundation → one private safe stream.
2. US1 → searchable correlated events.
3. US2 → protected destination configuration.
4. US3 → verified outage resilience.

## Notes

- All 18 tasks use the required checkbox, sequential ID, optional parallel marker, story label where applicable, and exact file paths.
- A direct Go Loki client, custom Grafana dashboard, and durable replay remain out of scope.

## Phase 7: Convergence

- [X] T019 Execute the default, deployment-disabled, and rejected protected-destination configuration scenarios defined in `specs/008-external-log-aggregation/quickstart.md` against `deployments/compose.yml`, recording the Alloy and backend diagnostics observed for each state and confirming Alloy's running state is the enable/disable signal per US2/AC1-AC3 (missing)
- [X] T020 Execute the 15-minute aggregation outage acceptance scenario in `specs/008-external-log-aggregation/quickstart.md` with Loki stopped and Alloy running, recording that every exercised publishing, authentication, and notification workflow kept its normal status and that `log-outage-recovered` was delivered after recovery per US3/AC1-AC2 (missing)
- [X] T021 Run `scripts/verify-request-latency.py` against the running Compose stack with an acceptance account and record the observed authenticated read p95 (must stay below 200 ms) and write p95 (must stay below 500 ms) results per T018 (partial)
- [X] T022 Correct the destination-restore step in `specs/008-external-log-aggregation/quickstart.md` so the local default is genuinely restored by unsetting `LOKI_URL`, `LOKI_TENANT_ID`, and `LOKI_BEARER_TOKEN_FILE` before recreating the Alloy service, since the current command reuses the rejected-destination shell overrides per T012 (partial)

## Phase 8: Convergence

- [X] T023 Record the completed 15-minute aggregation outage outcome in the `specs/008-external-log-aggregation/quickstart.md` acceptance-results section: iteration count, the status returned by each exercised workflow, the `log-outage-recovered` delivery after Loki restarted, the best-effort lost-event boundary, and whether the collector's local diagnostics reported the delivery failure within the outage window per US3/AC1-AC2 (partial)
- [X] T024 Resolve the tenant-identity logging conflict between `deployments/alloy/config.alloy` and `specs/008-external-log-aggregation/data-model.md`, which requires that `tenant_id` is never logged while Alloy's own delivery-failure diagnostics echo the configured tenant, by supplying the tenant through an `X-Scope-OrgID` header instead of the `tenant_id` endpoint field or by amending the data-model note to exempt collector component diagnostics per plan: delivery configuration (contradicts)
