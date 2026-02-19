# ADR-0001: Pulumi Schema to Kubernetes CRD Adapter (`pulumi-component-to-crd`)

## Document Scope
This ADR defines the CLI functionality contract and implementation decisions for `pulumi-component-to-crd`.

When updating this file:
- include user-visible CLI behavior, technical decisions, and accepted tradeoffs
- keep contracts explicit (flags, output channels, errors, exit codes, deterministic behavior)
- do not add contributor workflow, coding-style rules, or repository process guidance (those belong in `AGENTS.md`)

## Status
Accepted and implemented (current CLI baseline).

## Date
2026-02-18

## Context
Pulumi already produces package schema JSON (`resources`, `inputProperties`, `requiredInputs`, `types`).
This project intentionally does not re-implement source analysis. It adapts that schema into a single-version Kubernetes CRD YAML.

Primary constraints carried from planning:
- deterministic output
- single-line parseable errors (`error: ...`)
- single-line parseable skip warnings (`warn: ...`)
- stable exit-code contract (`0`, `2`, `3`, `4`)
- YAML written to `stdout`

## Decision
Build and keep a small adapter CLI with narrow package responsibilities:
- `cmd/component-to-crd`: orchestration + exit-code mapping
- `internal/schema`: load local path or HTTP(S) URL + preflight validation
- `internal/select`: deterministic component discovery and selection
- `internal/translate`: strict Pulumi-schema to OpenAPI translation
- `internal/emit`: identity derivation/validation + deterministic CRD YAML rendering

## Current CLI Contract (Implemented)
Command:
- `pulumi-component-to-crd --in <path-or-url> [--component <token>] [--group ... --version ... --kind ... --plural ... --singular ...] [--verbose]`

Behavior:
- `--in` is required; supports local files and `http(s)` URLs.
- `--component` is optional only when exactly one component exists.
- positional arguments are rejected.
- generated CRD YAML is written to `stdout` only.
- errors are emitted as parseable single-line `error: ...` entries on `stderr`.
- default mode does not emit `info:` or `warn:` diagnostics.
- `--verbose` emits additional stage logs (`info: ...`) and skipped-path diagnostics (`warn: ...`) on `stderr`.

Exit codes:
- `0`: success
- `2`: user/input/validation/preflight errors (including when no translatable top-level `spec` fields remain)
- `3`: hard unsupported translation construct (non-skippable)
- `4`: output serialization/write errors

## Implemented Technical Rules
Selection:
- discover `resources[*].isComponent == true`
- sort tokens deterministically
- selection rules:
  - explicit `--component` must exact-match
  - implicit auto-select when exactly one token
  - implicit zero/multiple is an error with sorted token list

Preflight validation (component-scoped, runs after selection):
- validates traversal shape for `properties`, `items`, `additionalProperties`
- validates local `$ref` targets (`#/types/...`, `#/resources/...`) exist and are traversable
- tolerates reachable non-local refs in selected-component mode (does not fail preflight solely on non-local prefix)
- malformed reachable local schema fails as `exit 2`

Translation (resilient):
- supports:
  - scalar types: `string`, `boolean`, `integer`, `number`
  - arrays (`items`)
  - objects (`properties`, `required`)
  - maps (`additionalProperties`)
  - local `$ref` resolution from `types` and `resources`
  - non-local `$ref` deterministic fallback:
    - `type: object`
    - `x-kubernetes-preserve-unknown-fields: true`
    - required field membership from `requiredInputs` is preserved
  - annotation passthrough:
    - `description` (string)
    - `default`
    - `enum`
- skips unsupported field paths (for example unsupported keywords/types such as `oneOf`, `anyOf`, `allOf`, `not`, `discriminator`, `patternProperties`, `const`) and, in `--verbose` mode, logs warnings to `stderr`
- preserves `exit 3` for intentionally non-skippable unsupported constructs
- required fields are normalized (deduplicated + sorted)
- required entries for skipped properties are removed to keep emitted schema valid
- fails as `exit 2` if all top-level `spec` properties are skipped and none remain translatable

CRD emission:
- deterministic key ordering and stable YAML text formatting
- emits a single CRD version (`served: true`, `storage: true`)
- always emits `scope: Namespaced`
- translated schema is wrapped under:
  - `spec.versions[0].schema.openAPIV3Schema.properties.spec`

Identity derivation and overrides:
- precedence: explicit flag > derived value > hard fallback constant
- derived fields:
  - `kind` from selected token type segment
  - `singular` from kebab-case `kind`
  - `plural` from simple pluralization (`+s`, or `+es` if already ending in `s`)
  - `group` from sanitized schema package name (or token package prefix) + `.components.platform`
  - `version` from schema major version when possible, otherwise `v1alpha1`
- final identity is validated (DNS-like group/version/name constraints and Kubernetes-style kind)

## Coverage Snapshot
Automated tests in this repository cover:
- CLI arg parsing, missing flags, positional-arg rejection
- file and URL input loading (including non-200 and timeout paths)
- deterministic selection behavior
- malformed-schema preflight and precedence over later stages
- translation success with deterministic skipped-path logging
- local `#/resources/...` ref preflight/translation success and unresolved-ref failures
- hard unsupported classification (`exit 3`) for non-skippable constructs
- selected-component preflight behavior (reachable vs unreachable defects)
- non-local ref fallback behavior and required-field preservation
- identity defaults, overrides, and validation failures
- deterministic output and output-writer failure (`exit 4`)

## Team Decisions (2026-02-18)
- Non-local refs use fallback schema: permissive object (`type: object` + `x-kubernetes-preserve-unknown-fields: true`).
- Fallback-backed fields remain in `required` when present in Pulumi `requiredInputs`.
- Non-local-ref provenance annotations are currently undecided and not required.
- Component-scoped preflight is the active contract: only reachable nodes from the selected component are validated before translation.
- Reachable composition keywords such as `oneOf` are currently handled by translation skip behavior with deterministic `warn:` output; if all top-level `spec` fields are skipped, the command returns `exit 2`.

## Explicitly Out of Scope (Current)
- Docker image packaging/distribution
- Kubernetes version-matrix validation gate (`1.32`-`1.35`)
- auth/retry/caching controls for URL input
- permissive translation mode (strict-only today)
- CRD `status` schema generation
- multi-version CRD emission
- operational flags like `--version`

## Consequences
Positive:
- clear adapter boundary, predictable failures, deterministic output
- small code surface and straightforward onboarding

Tradeoffs:
- strict unsupported policy can reject real-world schemas until mappings expand
- simple identity derivation heuristics prioritize determinism over linguistic correctness

This ADR is intended to be the living source of truth for the current CLI behavior.
