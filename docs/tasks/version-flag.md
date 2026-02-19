# Task 05: Add `--version` Flag for Binary Version Reporting

## Goal
Add a `--version` flag that prints the converter version and exits `0` without requiring `--in`.

## Why This Matters
Planning calls out version visibility for incident correlation. The current CLI has no way to report build/version metadata.

## Scope
In scope:
- Add `--version` boolean flag in `cmd/component-to-crd/main.go`.
- If `--version` is set, print version string to `stdout` and exit `0`.
- Do not require `--in` when `--version` is used.
- Support build-time injection via `-ldflags` (for example variable `version`).
- Keep existing behavior unchanged when `--version` is not set.

Out of scope:
- semantic version enforcement
- release automation

## Suggested Implementation
- Add package-level vars in `cmd/component-to-crd/main.go`:
  - `version = "dev"`
  - optionally `commit = ""`, `date = ""`
- Extend `parseArgs` config with `showVersion bool`.
- In `run`, short-circuit before schema load:
  - print version line
  - return `0`

## Acceptance Criteria
1. `pulumi-component-to-crd --version` exits `0` and prints a single line.
2. `pulumi-component-to-crd --version` does not require `--in`.
3. Existing `--in` flows and exit-code behavior remain unchanged.
4. Test coverage includes version short-circuit behavior.

## Validation
Automated:
- Add CLI tests in `cmd/component-to-crd/main_test.go`:
  - `--version` success without `--in`
  - `--version` ignores other missing-input conditions

Manual:
- Build binary with `./scripts/build_binary` and run `./bin/pulumi-component-to-crd --version`.

## Definition of Done
- Version output path implemented and tested.
- No behavior regressions for existing command paths.
