# Task 09: Add URL Schema Integration Tests for `init pulumi-component-promise`

## Outcome
Add deterministic command-level integration coverage for URL-based schema loading in `kratix init pulumi-component-promise`.

## Why This Slice Is Valuable
Task 02 added URL loading behavior and unit coverage in `internal/pulumi`, but command integration tests currently focus on local file schemas. This task closes that end-to-end coverage gap and protects CLI behavior around URL success/failure handling.

## Scope
In scope:
- Add integration tests that exercise `--schema` with URL values through the command path.
- Cover at minimum:
  - URL schema success path.
  - non-200 URL status failure path.
  - unreachable URL failure path.
- Keep tests deterministic in CI/sandboxed environments (no dependency on public internet).

Out of scope:
- Schema translation semantics beyond loader/selection behavior.
- Promise rendering/translation outputs from later tasks.

## File Touchpoints
- `test/init_pulumi_component_promise_test.go`
- Optional shared test helpers if needed under `test/`.

## Implementation Notes
- Avoid relying on external network connectivity.
- Prefer a test strategy that works in restricted/sandboxed environments (for example, local deterministic transport indirection or controlled server setup that is allowed by the harness).
- Keep assertions on stable error prefixes.

## Test Plan
Automated:
- Focused integration run:
```bash
go test ./test -count=1 -args -ginkgo.focus="init pulumi-component-promise"
```
- Broader sanity:
```bash
go test ./cmd/... ./internal/... ./test/... -count=1
```

## Acceptance Criteria
1. URL success and failure paths are covered at command integration level.
2. Tests are deterministic and do not require public internet access.
3. Assertions use stable error prefixes already established in loader code.
