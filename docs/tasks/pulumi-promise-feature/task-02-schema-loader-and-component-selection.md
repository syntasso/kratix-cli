# Task 02: Implement Pulumi Schema Loading and Component Selection

## Outcome
Build the schema input layer used by `init pulumi-component-promise`: local file/URL loading plus deterministic component selection with actionable errors.

## Why This Slice Is Valuable
This isolates the highest-risk UX area (input parsing + component disambiguation) and makes downstream translation/generation deterministic.

## Scope
In scope:
- Support `--schema` as:
  - local file path
  - HTTP/HTTPS URL
- Parse Pulumi package schema JSON into internal model.
- Select target component token:
  - if one component exists and `--component` not set: auto-select
  - if many components and `--component` missing: fail and list available tokens
  - if provided `--component` not found: fail and list available tokens
- Return clear errors for malformed JSON and unreachable URL.

Out of scope:
- OpenAPI translation details.
- Promise file rendering.

## Suggested Reuse
Leverage the shared Pulumi internals in the main CLI tree:
- `internal/pulumi/schema_loader.go`
- `internal/pulumi/component_select.go`

Preferred direction:
- Keep schema loading and component selection in `internal/pulumi/` and avoid shelling out to other binaries.

## File Touchpoints
- `cmd/init_pulumi_component_promise.go`
- `internal/pulumi/schema_loader.go` (new)
- `internal/pulumi/component_select.go` (new)
- `internal/pulumi/*_test.go` (new)
- `test/init_pulumi_component_promise_test.go` (extend)

## Implementation Steps
1. Create loader API: `LoadSchema(source string) (Document, error)`.
2. Create selector API: `SelectComponent(doc Document, token string) (Resource, error)`.
3. Normalize error messages with stable prefixes (for test assertions).
4. Wire loader/selector into command execution path after flag validation.
5. Pass selected component metadata forward (even if downstream is stubbed).

## Test Plan
Automated:
- Unit tests:
  - load local file success
  - load URL success via test HTTP server
  - malformed JSON error
  - non-200 URL status error
  - single-component auto-select
  - multi-component requires `--component`
  - unknown component error includes available tokens
- Integration tests:
  - command fails correctly for multi-component schema without `--component`
  - command succeeds when explicit component supplied
- Run:
```bash
go test ./internal/... ./cmd/... ./test/...
```

Manual smoke:
```bash
kratix init pulumi-component-promise demo --schema https://www.pulumi.com/registry/packages/eks/schema.json --component eks:index:Cluster
```

## Documentation Updates
- Add a short section in command docs or README explaining component selection behavior.
- Include error example showing token list when ambiguous.

## Acceptance Criteria
1. Local path and URL schema inputs both work.
2. Component selection behavior is deterministic and fully covered by tests.
3. Ambiguity and invalid component errors include available tokens.
4. Command path now reaches a selected component object for downstream steps.

## Definition of Done
- Loader + selector shipped with unit and integration tests.
- Teammates can reliably feed schema/component into later translation tasks.
- Docs explain how users resolve component-selection failures.
