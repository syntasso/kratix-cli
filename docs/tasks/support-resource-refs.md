# Task: Minimal `#/resources/...` Ref Support

## Problem
Running the CLI against real provider schemas (for example `pulumi-eks`) fails preflight with:

`unsupported ref "#/resources/..." (expected local ref prefix "#/types/")`

The current implementation only supports local type refs (`#/types/...`), which blocks valid schemas that use local resource refs in input shapes.

## Goal
Add minimal, tested support for local resource refs (`#/resources/...`) so `component-to-crd` can translate selected components that depend on resource-shaped inputs.

## Non-goals
- No support for remote refs.
- No support for `oneOf`/`anyOf`/`allOf` beyond existing behavior.
- No broad redesign of translation architecture.

## Scope
- `internal/schema/validate.go`
- `internal/translate/resolve.go`
- `internal/translate/translate_test.go`
- `internal/schema/validate_test.go`
- `cmd/component-to-crd/main_test.go` (single regression case)

## Proposed Minimal Behavior
1. Accept both local ref prefixes:
   - `#/types/<token>`
   - `#/resources/<token>`
2. Resolve `#/resources/<token>` by reading the resource definition from `Document.Resources`.
3. When translating a resource ref:
   - Treat the referenced resource input shape as an object schema rooted at the same translation rules used for component input properties.
   - Preserve deterministic ordering and required-field sorting.
4. Keep existing unsupported behavior unchanged for constructs still not handled.

## Error Contract
- Keep single-line parseable errors: `error: ...`.
- For unresolved refs, include component token + schema path context.
- Exit code mapping remains unchanged:
  - `2` user/input validation errors.
  - `3` unsupported translation constructs.
  - `4` output/serialization errors.

## Implementation Notes
1. Update schema preflight ref validation
   - Allow `#/resources/` in the same places `#/types/` is currently allowed.
   - For local resource refs, verify target exists in `doc.Resources`.
2. Extend ref resolver
   - Add a branch for `#/resources/<token>`.
   - Convert referenced resource input properties into an object schema in the resolver/translator path.
3. Guard against recursion
   - Reuse or add a simple visited-stack check for local ref cycles (`types` and `resources`) and emit actionable error context.

## Tests (Required)
1. `internal/schema/validate_test.go`
   - New passing case: resource ref to existing `#/resources/...`.
   - New failing case: unresolved `#/resources/...` with stable error text.
2. `internal/translate/translate_test.go`
   - New passing case: property `$ref` to `#/resources/...` emits expected object schema shape.
   - New failing case: cycle or unresolved resource ref returns contextual error.
3. `cmd/component-to-crd/main_test.go`
   - Regression test using a tiny schema fixture containing `#/resources/...` that previously failed preflight, now succeeds and emits stable YAML snippet.
   - Include deterministic output assertion (property/required ordering).

## Acceptance Criteria
- CLI can successfully process a schema containing local `#/resources/...` refs used by the selected component.
- Existing tests continue to pass.
- New tests cover success and failure paths for schema preflight + translation + CLI.
- Behavior remains deterministic for equivalent inputs.

## Verification
Run from `component-to-crd`:

```bash
go test ./...
```

