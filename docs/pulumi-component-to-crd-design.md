# ADR-0001: Pulumi Schema to Kubernetes CRD Adapter (`component-to-crd`)

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
- `component-to-crd --in <path-or-url> [--component <token>] [--group ... --version ... --kind ... --plural ... --singular ...]`

Behavior:
- `--in` is required; supports local files and `http(s)` URLs.
- `--component` is optional only when exactly one component exists.
- positional arguments are rejected.
- generated CRD YAML is written to `stdout` only.
- all user-facing errors are single-line `error: ...` on `stderr`.

Exit codes:
- `0`: success
- `2`: user/input/validation/preflight errors
- `3`: valid schema but unsupported translation construct
- `4`: output serialization/write errors

## Implemented Technical Rules
Selection:
- discover `resources[*].isComponent == true`
- sort tokens deterministically
- selection rules:
  - explicit `--component` must exact-match
  - implicit auto-select when exactly one token
  - implicit zero/multiple is an error with sorted token list

Preflight validation (runs before selection/translation):
- validates traversal shape for `properties`, `items`, `additionalProperties`
- validates `$ref` is local-only (`#/types/...`)
- validates referenced local type exists and is traversable
- malformed schema fails as `exit 2` before any selection or translation error

Translation (strict):
- supports:
  - scalar types: `string`, `boolean`, `integer`, `number`
  - arrays (`items`)
  - objects (`properties`, `required`)
  - maps (`additionalProperties`)
  - local `$ref` resolution from `types`
  - `enum` and `default` passthrough
- rejects unsupported keywords/types (for example `oneOf`, `anyOf`, `allOf`, `not`, `discriminator`, `patternProperties`, `const`) as `exit 3`
- required fields are normalized (deduplicated + sorted)

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
Automated tests in `component-to-crd` cover:
- CLI arg parsing, missing flags, positional-arg rejection
- file and URL input loading (including non-200 and timeout paths)
- deterministic selection behavior
- malformed-schema preflight and precedence over later stages
- translation success and unsupported classification (`exit 3`)
- identity defaults, overrides, and validation failures
- deterministic output and output-writer failure (`exit 4`)

## Explicitly Out of Scope (Current)
- Docker image packaging/distribution
- Kubernetes version-matrix validation gate (`1.32`-`1.35`)
- auth/retry/caching controls for URL input
- permissive translation mode (strict-only today)
- CRD `status` schema generation
- multi-version CRD emission
- operational flags like `--verbose` or `--version`

## Consequences
Positive:
- clear adapter boundary, predictable failures, deterministic output
- small code surface and straightforward onboarding

Tradeoffs:
- strict unsupported policy can reject real-world schemas until mappings expand
- simple identity derivation heuristics prioritize determinism over linguistic correctness

This ADR is intended to be the living source of truth for the current CLI behavior.
