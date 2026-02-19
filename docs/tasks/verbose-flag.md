# Task 06: Add `--verbose` Output Channels + Stage Logging

## Goal
Define a full output-mode contract where default runs write only to `stdout`, and `--verbose` enables `stderr` output (including stage logs and diagnostic lines).

## Why This Matters
Operators need clean, pipe-friendly default output, while still having an opt-in debug mode that exposes execution trace and diagnostics.

## Scope
In scope:
- Add `--verbose` boolean flag.
- Keep YAML output on `stdout` unchanged.
- In default mode (no `--verbose`), emit no `stderr` output.
- In verbose mode, emit stage logs and diagnostics to `stderr`.
- Keep diagnostic line formats parseable (`info: ...`, `warn: ...`, `error: ...`).

Out of scope:
- structured JSON logs
- per-property translation traces
- persistent log files

## Output Contract
### Default mode (no `--verbose`)
- Write normal command output to `stdout` only.
- Do not write to `stderr`.
- Maintain deterministic `stdout` for equivalent input.

### Verbose mode (`--verbose`)
- Keep all normal output on `stdout`.
- Also emit `stderr` diagnostics:
  - stage-level lifecycle logs (`info:`)
  - skipped-path diagnostics (`warn:`)
  - parseable errors (`error:`)

## Logging Contract
- Verbose stage lines must be prefixed with `info:`.
- Example stages:
  - `info: loading schema`
  - `info: preflight validation`
  - `info: selecting component`
  - `info: translating schema`
  - `info: rendering CRD`
- Warning lines must be `warn: ...`.
- Error lines must be `error: ...`.

## Acceptance Criteria
1. Default behavior writes only to `stdout`; `stderr` is empty.
2. With `--verbose`, `stderr` includes stage logs (`info:`) and existing diagnostics (`warn:`/`error:`) as applicable.
3. `stdout` output is byte-identical between verbose and non-verbose runs for the same successful input.
4. Diagnostics remain single-line parseable (`info:`, `warn:`, `error:`).
5. Exit code contract remains unchanged.
6. Tests verify both channel modes and deterministic `stdout`.

## Validation
Automated:
- CLI tests for:
  - successful run without `--verbose`: non-empty `stdout`, empty `stderr`
  - successful run with `--verbose`: same `stdout` bytes as default + expected `info:`/`warn:` on `stderr`
  - failing run without `--verbose`: parseable failure behavior on `stdout` only + expected exit code
  - failing run with `--verbose`: expected `error:` on `stderr` + expected exit code

Manual:
- Run converter with and without `--verbose`; confirm identical `stdout` and extra `stderr` only in verbose mode.

## Definition of Done
- `--verbose` implemented as output-channel gate plus stage-level logging.
- CLI tests cover channel behavior for success and failure paths.
- Existing deterministic output and exit-code contracts preserved.
