# Task 03: Translate Pulumi Component Inputs to CRD `spec` OpenAPI Schema

## Outcome
Implement schema translation from Pulumi component `inputProperties` into Kubernetes CRD OpenAPI under `openAPIV3Schema.properties.spec`.

## Why This Slice Is Valuable
This is the core API-contract conversion. Once done, generated Promises expose real typed request schemas.

## Scope
In scope:
- Translate selected component inputs into OpenAPI schema for `spec`.
- Preserve required fields.
- Support baseline preview type coverage:
  - string, number, integer, boolean
  - object properties
  - arrays
  - maps via `additionalProperties`
  - local refs where supported by existing translator
- Unsupported constructs are skipped with deterministic warnings (not panic).
- Fail if resulting `spec` is empty/unusable.

Out of scope:
- Exhaustive Pulumi feature parity.
- Stage output generation.

## Suggested Reuse
Build on the shared translator in:
- `internal/pulumi/translate.go`

## File Touchpoints
- `internal/pulumi/translate.go` (new or extracted)
- `internal/pulumi/translate_test.go` (new)
- `cmd/init_pulumi_component_promise.go` (wire into flow)
- `test/init_pulumi_component_promise_test.go` (assert translated fields in output once generation is wired)

## Implementation Steps
1. Define translation contract function, e.g.:
   - `TranslateInputsToSpecSchema(doc, component) (map[string]any, []string, error)`
2. Port/extract supported translation logic.
3. Add warning collector for skipped unsupported nodes.
4. Ensure deterministic output ordering where feasible (for stable snapshots).
5. Add guard for empty resulting `spec`.

## Test Plan
Automated:
- Unit tests for:
  - required propagation
  - nested objects/arrays/maps
  - ref resolution for supported local refs
  - unsupported node skipped with warning
  - empty translated schema returns error
- Run:
```bash
go test ./internal/...
```

Manual validation:
- Run command with known schema fixture and inspect emitted Promise API after generation is available.

## Documentation Updates
- Add translator support matrix (supported vs skipped constructs) to `docs/pulumi-promise-feature.md` or command-specific doc.

## Acceptance Criteria
1. Translated OpenAPI schema is generated for supported Pulumi inputs.
2. Required fields are preserved.
3. Unsupported constructs are skipped with stable warnings.
4. Empty-result scenarios fail with clear errors.
5. Unit tests cover mappings and error/warning behavior.

## Definition of Done
- Translator implementation is production-usable for preview scope.
- Tests prevent regressions in supported type mapping.
- Behavior for unsupported schema constructs is explicit and documented.
