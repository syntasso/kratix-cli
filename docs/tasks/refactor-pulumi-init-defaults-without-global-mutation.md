# Task 10 (Optional): Resolve Pulumi Init Defaults Without Mutating Global Flags

## Outcome
Refactor `init pulumi-component-promise` to compute defaulted `version` and `plural` values locally, without mutating shared package-level state.

## Why This Slice Is Valuable
Task 04 introduced in-command defaulting by writing to shared globals. That pattern works, but it increases coupling and makes behavior harder to reason about for in-process command execution and future tests.

## Scope
In scope:
- `cmd/init_pulumi_component_promise.go` default resolution for `version` and `plural`.
- Any small helper functions needed to keep call sites clear.
- Tests that assert behavior remains unchanged.

Out of scope:
- Migrating every `init` command (covered by broader Task 08).
- User-facing CLI contract changes.

## Implementation Contract
1. Keep generated output and CLI behavior identical.
2. Replace global mutation with local derived values, for example:
   - `resolvedVersion := defaultVersion(version)`
   - `resolvedPlural := defaultPlural(plural, kind)`
3. Pass resolved values through CRD and README flag reconstruction paths explicitly.
4. Keep changes small and easy for a junior engineer to follow.

## Test Plan
Automated:
```bash
go test ./cmd/...
go test ./test/... -ginkgo.focus="init pulumi-component-promise"
```

Assertions:
- Existing flat/split fixture tests remain green.
- Existing README reconstruction tests remain green.
- No user-visible error/help text changes.

## Acceptance Criteria
1. `init pulumi-component-promise` no longer assigns to shared `version`/`plural` globals.
2. Output artifacts are unchanged for existing test cases.
3. Tests covering pulumi init generation and README reconstruction pass.
