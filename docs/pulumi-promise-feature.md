# Pulumi Component Promise Init (Preview)

## Summary
`kratix init pulumi-component-promise` bootstraps a Promise from a Pulumi package schema.

```bash
kratix init pulumi-component-promise PROMISE-NAME \
  --schema PATH_OR_URL \
  --group API-GROUP \
  --kind API-KIND \
  [--component TOKEN] [--version VERSION] [--plural PLURAL] [--split] [--dir DIR]
```

The command:
1. Loads a Pulumi package schema from a local path or HTTP(S) URL.
2. Selects a Pulumi component deterministically.
3. Translates component `inputProperties` into Promise API `spec`.
4. Writes standard Promise output (flat or split mode).
5. Adds a resource `configure` workflow container for Pulumi program generation.

## Current Implementation
- Command entrypoint: `cmd/init_pulumi_component_promise.go`
- Schema loading: `internal/pulumi/schema_loader.go`
- Component selection: `internal/pulumi/component_select.go`
- Translation: `internal/pulumi/translate.go`
- Translation helpers:
  - `internal/pulumi/translate_refs.go`
  - `internal/pulumi/translate_fields.go`
  - `internal/pulumi/translate_annotations.go`
  - `internal/pulumi/translate_warnings.go`

## UX Contract
- `--schema` is required.
- `--component` is optional only when the schema has exactly one component.
- If multiple components exist and `--component` is omitted, the command fails and lists valid component tokens.
- If `--component` is unknown, the command fails and lists valid component tokens.
- Preview warning is printed on execution paths (not on `--help`).

## Translation Coverage (Preview)
| Pulumi construct | Status | Behavior |
| --- | --- | --- |
| `string`, `number`, `integer`, `boolean` | Supported | Mapped directly to OpenAPI `type`. |
| `object` + `properties` | Supported | Translated recursively. |
| `required` | Supported | Preserved for translated properties. |
| `array` + `items` | Supported | `items` translated recursively. |
| `additionalProperties` | Supported | Mapped to OpenAPI map schema. |
| Local `$ref` (`#/types/...`, `#/resources/...`) | Supported | Resolved recursively; cycles fail with clear error. |
| Non-local `$ref` | Supported (fallback) | Permissive object with `x-kubernetes-preserve-unknown-fields: true`. |
| `oneOf`, `anyOf`, `allOf`, `not`, `discriminator`, `patternProperties`, `const` | Skipped with warning | Omitted from translated output and surfaced as deterministic warning text. |

## Workflow Contract
Generated resource `configure` workflow uses:
- Container name: `from-api-to-pulumi-pko-program`
- Container image: `ghcr.io/syntasso/kratix-cli/from-api-to-pulumi-pko-program:v0.1.0`
- Env vars:
  - `PULUMI_COMPONENT_TOKEN`
  - `PULUMI_SCHEMA_SOURCE`

After generation, teams can edit these env vars directly in:
- flat output: `promise.yaml` under `spec.workflows.resource.configure[].spec.containers[].env`
- split output: `workflows/resource/configure/workflow.yaml` under `spec.containers[].env`

## Stage Runtime Contract
The Pulumi stage runtime lives at `stages/pulumi-promise`.

Inputs:
- `KRATIX_INPUT_FILE` (optional, default `/kratix/input/object.yaml`)
- `PULUMI_COMPONENT_TOKEN` (required)

Output:
- `KRATIX_OUTPUT_FILE` (optional, default `/kratix/output/object.yaml`)

Transformation:
- Reads one Kratix request object.
- Requires `metadata.name` and `spec` (where `spec` must be an object).
- Emits one `pulumi.com/v1` `Program` manifest.
- Copies request `metadata.labels` and `metadata.annotations`.
- Preserves request namespace (defaults to `default` if unset).
- Maps request `spec` to `spec.resources.<component-token>.properties`.
- Uses deterministic Program naming: `<request-name>-<hash>`.

## Test Coverage And Fixtures
- Command integration tests: `test/init_pulumi_component_promise_test.go`
- Unit tests:
  - `internal/pulumi/schema_loader_test.go`
  - `internal/pulumi/component_select_test.go`
  - `internal/pulumi/translate_test.go`
- Test schema fixture: `test/assets/pulumi/schema.valid.json`
- Expected generated output snapshots:
  - `test/assets/pulumi/expected-output/promise.yaml`
  - `test/assets/pulumi/expected-output/example-resource.yaml`
  - `test/assets/pulumi/expected-output/README.md`
  - `test/assets/pulumi/expected-output-with-split/api.yaml`
  - `test/assets/pulumi/expected-output-with-split/dependencies.yaml`
  - `test/assets/pulumi/expected-output-with-split/example-resource.yaml`
  - `test/assets/pulumi/expected-output-with-split/README.md`
  - `test/assets/pulumi/expected-output-with-split/workflows/resource/configure/workflow.yaml`

Run focused tests:

```bash
go test ./internal/pulumi/... ./test/... -count=1
```
