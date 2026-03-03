# Feature Request: Add Pulumi Stack Emission via a Second Stage Entry Point

## Outcome
Extend `stages/pulumi-promise` to emit a runnable PKO `Stack` resource that references the already-generated `Program`, using a separate stage entry point and a second workflow container.

## Dependency
This feature builds on a successful implementation of `Feature Request: Extend Pulumi Promise Stage to Emit a Working PKO Program`.

## Problem Statement
Current scope delivers a valid PKO `Program`, but users still need manual steps to create and wire a `Stack` for execution.

We need a clear follow-on feature that keeps stage logic in the same codebase to keep low maintenance costs.

## Proposed Scope (Small Slices)
### Slice 1: Add a second stage entry point in the same stage codebase
1. Add a second stage executable entry point under `stages/pulumi-promise` dedicated to emitting `Stack` resources.
2. Reuse shared helpers and contracts from the existing `Program` stage where practical.
3. Keep the current entry point unchanged to preserve existing behaviour.
   - Update generated workflow configuration so stack emission runs in a second container.
   - Do not extend or overload the existing program stage container.
   - Keep ordering and hand-off deterministic so the stack container runs after program output is available.

### Slice 2: Emit deterministic PKO Stack data only
When making this change, treat this like a bug fix
1. Emit `apiVersion: pulumi.com/v1`, `kind: Stack`.
2. Set `spec.programRef.name` to the deterministic `Program` name produced by the program stage.
3. Emit only fields that are fully deterministic from the existing request + stage contracts:
   - deterministic metadata/name/namespace behaviour already defined by stage contracts,
   - deterministic metadata passthrough already defined by stage contracts,
   - `spec.programRef.name` from the existing deterministic Program naming contract.
4. Treat `schema.json` + request `spec` as insufficient for operator-intent Stack runtime fields, so these must not be auto-generated.
5. Do not infer or default stack runtime intent:
   - do not guess auth/env reference wiring,
   - do not guess policy or runtime control fields.

### Slice 3: Deterministic stack identity
1. Set `spec.stack` from deterministic stage inputs only.
2. Validate deterministic field generation early, and fail with explicit, actionable errors when required request metadata is missing.
3. Do not apply implicit defaults for operator-intent fields.

### Slice 4: Init/readme documentation aligned with existing style
1. Update `kratix init pulumi-component-promise` generated docs to describe:
   - two-container workflow model (`Program` container + `Stack` container),
   - separate stage entry points in the same stage codebase,
   - deterministic Stack fields that are auto-generated,
   - an alternative option of writing a custom stage to update generated files before writing to outputs.
2. Match existing patterns and tone in current Pulumi docs and keep wording in British English.

## Non-Goals
- Reworking Pulumi schema translation in `internal/pulumi`.
- Changing the existing `Program` container behaviour or contract.
- Combining Program and Stack emission into one overloaded container path.

## Test Plan
Automated tests in `stages/pulumi-promise/test` should cover:
1. New entry point emits valid `Stack` with required PKO fields.
2. `spec.programRef.name` deterministically matches Program naming contract.
3. Deterministic slice emits only deterministic Stack fields and does not populate operator-intent fields.
4. Existing Program entry point tests remain unchanged and green.

CLI init tests should also cover:
1. Generated workflow includes a second container for stack emission.
2. Generated README explains the two-container model and deterministic stack field generation without brittle exact-text assertions.

Run via:
```bash
make test-pulumi-promise-stage
```
and ensure top-level integration remains green:
```bash
go test ./test/... -ginkgo.focus="init pulumi-component-promise"
```

## Acceptance Criteria
1. A new stage entry point in `stages/pulumi-promise` emits PKO `Stack` resources.
2. Workflow generation uses a second container for stack emission instead of extending the existing container.
3. Deterministic Stack output references the Program and emits only contract-safe deterministic fields.
4. Stack field generation remains deterministic from existing stage inputs and contracts.
5. Program stage behaviour remains backward compatible.
6. Tests and generated docs are updated and pass, with junior-friendly clarity, and generated docs include custom-stage extension options.

## Sources
- Pulumi Kubernetes Operator API docs (Program/Stack types): https://pkg.go.dev/github.com/pulumi/pulumi-kubernetes-operator/v2/operator/api/pulumi/v1
- PKO Stacks docs: https://raw.githubusercontent.com/pulumi/pulumi-kubernetes-operator/master/docs/stacks.md
