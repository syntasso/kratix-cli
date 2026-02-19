# Task: Component-Scoped Preflight Validation

## Problem
Current preflight validates refs and schema node shapes across all resources and types before component selection/translation completes.

This causes UX issues:
- Users are blocked by malformed or unsupported refs in unrelated components/resources.
- Error messages are precise but can feel irrelevant to the requested `--component`.

## Goal
Keep the benefits of preflight (deterministic, path-aware malformed-input errors) while limiting validation scope to schema nodes reachable from the selected component.

## Field Validation Against Real Schema
Running:

```bash
go run ./cmd/component-to-crd \
  --in https://raw.githubusercontent.com/pulumi/pulumi-eks/master/provider/cmd/pulumi-resource-eks/schema.json \
  --component eks:index:Cluster
```

currently fails on a reachable cross-package ref:

`/aws/v7.14.0/schema.json#/types/aws:eks%2FAccessPolicyAssociationAccessScope:AccessPolicyAssociationAccessScope`

This ref is in the selected component graph, so the current behavior still blocks a practical CLI use case.

After non-local-ref fallback is applied, the same command still fails later in translation with:

`component "eks:index:Cluster" path "spec.fargate" unsupported construct: keyword "oneOf"`

This is also in the selected component graph and currently blocks successful CRD generation for this real schema.

## Desired UX State
- If `--component` is selected (explicitly or auto-selected), preflight validates only nodes reachable from that component’s input graph.
- Unrelated schema defects do not block translation for a valid selected component.
- Errors remain single-line, parseable, and path-aware.
- Reachable non-local refs do not hard-fail preflight when local traversal can continue safely.
- For reachable non-local refs, translation emits a deterministic fallback schema so CRD generation can succeed for common provider schemas.
- Reachable union/composition keywords currently marked unsupported (starting with `oneOf`) can be translated via deterministic fallback so this `eks:index:Cluster` flow succeeds.

## Non-goals
- No change to unsupported construct handling in translation (`exit 3`).
- No broad rewrite of selection/translation architecture.

## Functional Requirements
1. Select component first, then run preflight in selected-component scope.
2. Traverse reachable nodes from selected component input properties through:
   - local type refs: `#/types/...`
   - local resource refs: `#/resources/...`
   - reachable non-local refs (for detection + fallback handling)
   - nested object/array/additionalProperties nodes already supported by preflight
3. Validate only reachable nodes:
   - malformed reachable nodes fail with `exit 2`
   - malformed unreachable nodes do not fail the run
4. Reachable non-local refs:
   - do not fail preflight as malformed input solely due to non-local prefix
   - are translated with a deterministic permissive fallback schema (for example object-preserving-unknown-fields) rather than failing
5. Reachable union/composition keywords (at minimum `oneOf`) in selected-component scope:
   - do not return unsupported construct errors when fallback mode is applicable
   - map to deterministic permissive fallback schema
6. Preserve deterministic traversal/order for equivalent inputs.

## Exit/Error Contract
- Keep existing exit code contract:
  - `2` user/input errors (including malformed reachable schema)
  - `3` unsupported translation constructs
  - `4` output/serialization failures
- Keep error format: `error: ...` (single line)
- Keep actionable context: include selected component token + schema path when possible.

## Proposed Implementation
1. Introduce component-scoped preflight entrypoint
   - Example: `ValidateForTranslationComponent(doc, componentToken string) error`
2. Refactor validator traversal
   - Seed traversal from selected resource input properties
   - Follow refs lazily as discovered
   - Track visited refs/nodes to avoid cycles and duplicate work
3. Add non-local ref strategy used by both preflight and translation
   - classify refs as local type/resource vs non-local
   - keep local refs strict (must resolve)
   - map non-local refs to deterministic fallback schema during translation
4. Add targeted unsupported-keyword fallback in translation (start with `oneOf`)
   - apply only for selected-component traversal path where direct support is unavailable
   - preserve deterministic output shape and required field handling
5. Update CLI orchestration in `cmd/component-to-crd/main.go`
   - Order:
     1) load schema
     2) discover/select component
     3) component-scoped preflight
     4) translate
6. Keep existing full-document validator only if needed by tests/internal callers; otherwise replace with scoped path.

## Tests (Required)
1. `internal/schema/validate_test.go`
   - Reachable malformed ref fails.
   - Unreachable malformed ref does not fail scoped validation.
   - Deterministic behavior for equivalent graph shapes/order.
2. `cmd/component-to-crd/main_test.go`
   - Regression: schema contains unrelated bad ref in non-selected component; selected component still succeeds.
   - Regression: bad ref in selected component still fails with `exit 2` and preflight path.
   - Precedence: reachable preflight errors still win over translation/selection follow-on errors where applicable.
   - Regression: selected component with reachable non-local ref no longer fails preflight.
   - Regression: selected component with reachable `oneOf` path uses fallback and no longer exits `3` for that path.
3. Existing translation tests remain green.
4. Add a focused translation unit test that verifies deterministic fallback rendering for non-local refs.
5. Add a focused translation unit test for deterministic fallback rendering of `oneOf`.

## Acceptance Criteria
- Running against schemas like `pulumi-eks` is not blocked by unrelated malformed refs outside the selected component graph.
- Reachable malformed refs still fail early with clear preflight errors.
- Running the `eks:index:Cluster` example above succeeds and emits CRD YAML.
- All tests pass with deterministic output unchanged for successful runs.

## Verification
Run from `component-to-crd`:

```bash
go test ./...
```
