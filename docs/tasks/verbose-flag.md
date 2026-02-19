# Task 06: Add `--verbose` Stage Logging

## Goal
Add optional `--verbose` logging to `stderr` for stage-level execution visibility while preserving parseable error lines.

## Why This Matters
Planning calls out debug visibility. Current failures are concise but there is no execution trace for operators.

## Scope
In scope:
- Add `--verbose` boolean flag.
- Emit stage logs to `stderr` only when enabled.
- Keep error line format unchanged (`error: ...`).
- Keep YAML output on `stdout` unchanged.

Out of scope:
- structured JSON logs
- per-property translation traces
- persistent log files

## Logging Contract
- Verbose lines should be prefixed with `info:`.
- Example stages:
  - `info: loading schema`
  - `info: preflight validation`
  - `info: selecting component`
  - `info: translating schema`
  - `info: rendering CRD`
- Error lines remain exactly `error: ...`.

## Acceptance Criteria
1. Default behavior (without `--verbose`) is byte-identical to current stdout/stderr for existing tests.
2. With `--verbose`, stage logs are emitted to `stderr` and command behavior is otherwise unchanged.
3. Errors remain single-line `error: ...` and continue to use existing exit codes.
4. Tests verify verbose and non-verbose paths.

## Validation
Automated:
- CLI tests for:
  - successful run with `--verbose` includes `info:` stages
  - failure with `--verbose` includes `info:` and one `error:` line
  - default mode unchanged

Manual:
- Run converter with and without `--verbose` and compare stdout identity.

## Definition of Done
- `--verbose` implemented with stage-level logging and tests.
- Existing contracts preserved.
