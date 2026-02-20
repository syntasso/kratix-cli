# Pulumi Promise Init (Preview) - Feature Design

## Summary
Add a new preview command:

```bash
kratix init pulumi-component-promise PROMISE-NAME --schema PATH_OR_URL [--component TOKEN] [--group GROUP] [--kind KIND] [--version VERSION] [--plural PLURAL] [--split] [--dir DIR]
```

The command should:
1. Read a Pulumi package schema JSON.
2. Convert the selected Pulumi component inputs into the Promise CRD `spec.api` schema.
3. Generate a Kratix resource `configure` workflow stage that transforms Promise requests into a Pulumi Kubernetes Operator Program resource.
4. Produce output matching existing init flows (`promise.yaml` or split files, `example-resource.yaml`, `README.md`).

This feature is preview-only in the same way as `helm-promise`, `crossplane-promise`, and `operator-promise`.

## Why This Exists
Current `kratix init` flows cover Terraform modules, Helm charts, Crossplane XRDs, and Operators. Pulumi users need equivalent scaffolding to go from existing IaC contract/schema to a usable Promise without hand-authoring CRD and workflow manifests.

## Existing Codebase Touchpoints

### Command entrypoints and patterns
- `cmd/init.go`: shared persistent flags (`--group`, `--kind`, `--version`, `--plural`, `--dir`, `--split`).
- `cmd/init_helm_promise.go`: preview command structure and chart-to-schema flow.
- `cmd/init_operator_promise.go`: reusable helpers for full Promise object generation (`getFilesToWrite`, `writePromiseFiles`, `generateResourceConfigurePipelines`).
- `cmd/init_crossplane_promise.go`: preview messaging + destination selectors + API translation pattern.
- `cmd/preview_warning.go`: current preview warning behavior (`printPreviewWarning`).

### Templates and generated docs
- `cmd/templates/promise/*.tpl`
- `cmd/templates/promise/README.md.tpl` (contains reconstruction command for how Promise was generated).

### Stage implementations
- Existing stage binaries:
  - `stages/operator-promise/main.go`
  - `stages/crossplane-promise/main.go`
  - `stages/terraform-module-promise/main.go`
- Existing Pulumi conversion prototype:
  - `stages/pulumi-component-promise/pulumi-component-to-crd/...`
  - Includes CLI, translation pipeline, tests, and design notes.

### Test harnesses
- Integration-style command tests under `test/`:
  - `test/init_helm_promise_test.go`
  - `test/init_operator_promise_test.go`
  - `test/init_tf_module_promise_test.go`
  - `test/init_crossplane_promise_test.go`
- Shared expectations around generated files, workflow contents, schema content, and README command lines.

### Release plumbing
- Stage components are separately versioned in `release-please-config.json` and `.release-please-manifest.json`.
- Each stage has its own `stages/<stage>/CHANGELOG.md`, `Dockerfile`, and `Makefile`.

## Proposed User Experience

### Command
```bash
kratix init pulumi-promise PROMISE-NAME \
  --schema https://.../schema.json \
  --component pkg:index:MyComponent \
  --group syntasso.io \
  --kind Database
```

### Required flags
- `--schema`: Pulumi package schema source (local file path or URL).

### Optional flags
- `--component`: Pulumi component token. Required if schema has multiple component resources.
- `--destination-selector key=value` (optional extension; if omitted, use sensible default labels for Pulumi destinations).

### Reused global flags
- `--group`, `--kind`, `--version`, `--plural`, `--dir`, `--split` from `cmd/init.go`.

### Preview behavior
- Mark command as preview in `Short` and `Long` help text.
- Call `printPreviewWarning()` at runtime like existing preview commands.

## Functional Scope

### In scope
1. Convert Pulumi component input schema into Promise CRD `spec.versions[0].schema.openAPIV3Schema.properties.spec`.
2. Create resource configure workflow pipeline with one container that transforms request input into Program CR.
3. Include workflow env that identifies target Pulumi stack/project settings and selected component identity.
4. Generate standard bootstrap files (`promise.yaml` or split representation, `README.md`, `example-resource.yaml`).
5. Maintain deterministic output and useful error messages for invalid schema/component selection.

### Out of scope for first preview
- Full parity with every Pulumi schema feature if translator cannot represent all constructs.
- Promise configure/dependency workflow generation for Pulumi operator install (can be documented as manual follow-up or future flag).
- Automatic provider credentials/secret wiring.
- Multi-component generation in one command.

## Design Proposal

### 1) Add `init pulumi-component-promise` command
Create `cmd/init_pulumi_component_promise.go` with a shape matching other init subcommands:
- Cobra command registration in `init()`.
- Flag parsing and validation.
- `RunE: InitPulumiComponentPromise`.

Suggested command constants:
- Container name: `from-api-to-pulumi-pko-program`.
- Container image: `ghcr.io/syntasso/kratix-cli/from-api-to-pulumi-pko-program:<version>`.

### 2) Schema translation strategy
Preferred approach: reuse the existing translator prototype already in repo at:
- `stages/pulumi-component-promise/pulumi-component-to-crd`

Two implementation paths:
1. Fastest: wrap/execute translator binary behavior in command path (less ideal coupling).
2. Better: copy or extract translation internals into a reusable Go package in this repo (e.g. `internal/pulumi_schema.go`) and call directly from `InitPulumiPromise`.

Recommendation: path (2), to keep `kratix init` self-contained and testable without shelling out. Be sure to clean up the code in the current stages directory once all features are extracted into the kratix init CLI code.

Minimum translation contract for preview:
- Select component token deterministically (single auto-select or explicit `--component`).
- Preserve required fields and basic OpenAPI-compatible types.
- Skip unsupported paths with deterministic warnings rather than panic.
- Fail with clear message if no translatable top-level `spec` fields remain.
- Make sure to default all versions to v1alpha1 unless user provides an override.

Translator support matrix (preview):

| Pulumi schema construct | Status | Behavior |
| --- | --- | --- |
| `type: string`, `number`, `integer`, `boolean` | Supported | Mapped directly to OpenAPI `type`. |
| `type: object` + `properties` | Supported | Properties translated recursively. |
| `required` | Supported | Preserved and filtered to translated properties only. |
| `type: array` + `items` | Supported | `items` translated recursively. |
| `additionalProperties` | Supported | Mapped as OpenAPI map values schema. |
| Local `$ref` (`#/types/...`, `#/resources/...`) | Supported | Resolved recursively; cycles fail with a clear error. |
| Non-local `$ref` | Supported (fallback) | Converted to permissive object with `x-kubernetes-preserve-unknown-fields: true`. |
| `oneOf`, `anyOf`, `allOf`, `not`, `discriminator`, `patternProperties`, `const` | Skipped | Path is omitted and surfaced as deterministic warning text. |

### 3) Promise API generation
Use existing promise generation flow for consistency:
- Build CRD object with `group/kind/version/plural` from flags/defaults.
- Put translated Pulumi schema under `spec` properties.
- Reuse `getFilesToWrite(...)` and `writePromiseFiles(...)` from `cmd/init_operator_promise.go`.

### 4) Resource configure stage generation
Generate pipeline via `generateResourceConfigurePipelines(...)` with env vars like:
- `PULUMI_COMPONENT_TOKEN`
- `PULUMI_SCHEMA_SOURCE` (optional, mainly for traceability)
- `PULUMI_STACK` / `PULUMI_PROJECT` (if user-facing flags are included now)
- `PROGRAM_NAMESPACE` defaulting to request namespace if not set

Stage responsibility:
- Read Kratix input request (`/kratix/input/object.yaml`).
- Map `spec` from Promise request into Pulumi Program CR payload.
- Emit Program CR YAML to `/kratix/output`.

### 5) Pulumi Kubernetes Operator Program CR mapping
Expected high-level mapping:
- Request metadata drives Program resource metadata (`name`, `namespace`, labels for traceability).
- Request `spec` becomes Program config/variables section used by selected component program.
- Ensure idempotent naming to avoid collisions (e.g. `<kind>-<namespace>-<name>`).
- Ensure a specific Program CR schema/version for operator target in docs and code (must lock to a tested operator release).

## Usability and Documentation Requirements

### CLI help and examples
Add concrete examples in command `Example` string:
- Local schema file with explicit component.
- URL schema with auto-selected single component.
- Split mode.

### Generated README quality
Ensure generated `README.md` includes:
- Exact `kratix init pulumi-promise ...` command used.
- How to build/push the Pulumi stage image if custom image is needed.
- How to inspect/edit destination selectors.
- Pulumi operator prerequisites.

### Top-level docs
After command exists, update:
- `README.md` usage section.
- `docs/design.md` with new `init from pulumi` stanza.

## Testing Strategy

### CLI integration tests (`test/`)
Add `test/init_pulumi_promise_test.go` covering:
1. Required flags and argument validation.
2. Multiple components requiring `--component`.
3. Successful generation with `--split` and non-split.
4. Expected files and README command reconstruction.
5. Workflow pipeline container image, name, and env vars.
6. CRD schema contains translated Pulumi fields and required entries.
7. Failure mode for invalid schema input.

### Translator unit tests
If translator code is moved to `internal/`:
- Add focused tests for type mappings, required handling, unsupported path skipping, deterministic behavior.

### Stage tests (`stages/pulumi-promise`)
Mirror other stages:
- `test/stage_suite_test.go`
- `test/stage_test.go`

Validate:
- Input request => output Program CR correctness.
- Metadata mapping and naming stability.
- Error behavior on malformed inputs.

## Implementation Plan (Phased)

### Phase 1: Command and API translation
- Add `cmd/init_pulumi_promise.go`.
- Wire schema translation into CRD generation.
- Generate Promise files using existing helper path.
- Add integration tests with fixture schema(s).

### Phase 2: Stage implementation
- Add `stages/pulumi-promise` with Dockerfile/Makefile/main/test.
- Add container image reference in command pipeline generation.
- Verify stage output against Pulumi Program CR fixture.

### Phase 3: Docs and UX hardening
- Improve examples, error strings, README template behavior.
- Update root docs (`README.md`, `docs/design.md`).

### Phase 4: Release readiness
- Add stage component to release automation configs:
  - `release-please-config.json`
  - `.release-please-manifest.json`
- Add `stages/pulumi-promise/CHANGELOG.md`.
- Smoke test local build and stage image build commands.

## Release Plan

### Versioning and changelog
- CLI change should ship under a `feat:` conventional commit to appear in root changelog.
- Stage image should be independently releasable like existing stage components.

### Suggested rollout
1. Merge behind preview command label and warning.
2. Publish stage image with explicit preview tag (`v0.1.0`).
3. Announce as preview in release notes with supported Pulumi schema constraints.
4. Collect feedback on schema coverage before GA hardening.

## Risks and Mitigations

1. Pulumi schema complexity exceeds current translator support.
- Mitigation: deterministic skip warnings + explicit unsupported docs + fixture coverage for known packages.

2. Drift with Pulumi Kubernetes Operator Program CRD version.
- Mitigation: pin operator API version and document tested matrix.

3. Poor first-run UX due to ambiguous component selection.
- Mitigation: clear error listing available component tokens.

4. Runtime stage failures are hard to debug.
- Mitigation: add concise, parseable stage logs and include trace labels in emitted Program CR.

## Open Decisions

1. Should we expose stack/project/organization flags in `init`, or only through editable workflow env after generation?
  - Only through the editable environment variables after creation for now. This will reduce the complexity of the CLI. But make sure to document this optionality for users.
2. Do we want default destination selectors for Pulumi (for example `environment: pulumi`) like Terraform’s convention?
  - Yes
3. Should unsupported Pulumi schema fields fail fast or be skipped in preview by default?
  - Skipped but reported on. Matching the behaviour in terraform.
4. Should existing `stages/pulumi-component-promise/pulumi-component-to-crd` be promoted/reused directly or subsumed into shared internal packages?
  - subsumed.

## Definition of Done for Preview

1. `kratix init pulumi-component-promise ...` works for at least one real Pulumi package schema and one local fixture.
2. Generated Promise includes valid CRD schema, workflow, README, and example resource.
3. Generated workflow stage emits valid Program CR for Pulumi Kubernetes Operator.
4. New tests pass in CI (command integration + stage tests).
5. Release automation includes new stage component and changelog updates.
6. Documentation updates are merged for command usage and preview caveats.
