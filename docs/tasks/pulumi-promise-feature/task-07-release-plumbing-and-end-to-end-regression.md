# Task 07: Release Plumbing, Docs Hardening, and End-to-End Regression Coverage

## Outcome
Finalize preview readiness by wiring release metadata, completing user docs, and adding end-to-end regression coverage for the full command and stage flow.

## Why This Slice Is Valuable
This ensures the feature is test-complete, documented, and shippable as preview rather than only locally functional.

## Scope
In scope:
- Add new stage component to release automation:
  - `release-please-config.json`
  - `.release-please-manifest.json`
- Ensure stage `CHANGELOG.md` and versioning conventions are in place.
- Update top-level docs:
  - `README.md`
  - `docs/design.md`
- Add/expand end-to-end regression tests validating generated Promise and stage behavior together.

Out of scope:
- GA hardening or removal of preview label.

## Dependencies
Requires Tasks 01-06 to be complete.

## File Touchpoints
- `release-please-config.json`
- `.release-please-manifest.json`
- `README.md`
- `docs/design.md`
- `test/init_pulumi_component_promise_test.go` (or dedicated e2e test)
- stage regression fixtures if needed

## Implementation Steps
1. Register `stages/pulumi-promise` as releasable component mirroring existing stage entries.
2. Add/verify changelog bootstrap for stage.
3. Update root docs with:
  - command syntax
  - preview caveats
  - prerequisites (Pulumi K8s Operator)
  - split/non-split usage examples
4. Add regression test that covers end-to-end preview path:
  - init command output
  - expected workflow references stage image/env
  - stage consumes representative request and emits expected Program CR
5. Run full relevant test suites.

## Test Plan
Automated:
```bash
go test ./test/... ./stages/pulumi-promise/...
```

If feasible, include one higher-level regression test invoking the binary and stage fixtures.

Manual checks:
- Verify release config entries are syntactically valid and consistent with other components.
- Verify docs command examples run unchanged.

## Acceptance Criteria
1. Release automation includes the new stage component.
2. User-facing docs are complete for preview usage and prerequisites.
3. Regression tests cover full happy path across init output and stage transformation.
4. All relevant tests pass in CI.

## Definition of Done
- Feature can be merged and released in preview with confidence.
- A teammate can follow docs and run the full flow without internal tribal knowledge.
- Release metadata and changelog setup are complete.
