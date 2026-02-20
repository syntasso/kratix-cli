# Task 05: Add Resource Configure Workflow Scaffolding for Pulumi Program CR Output

## Outcome
Generated Promises include a resource `configure` workflow pipeline container that transforms request input into a Pulumi Kubernetes Operator Program CR.

## Why This Slice Is Valuable
This makes generated Promise not just declarative schema, but operationally usable.

## Scope
In scope:
- Add resource configure pipeline with one container:
  - container name: `from-api-to-pulumi-pko-program`
  - image: `ghcr.io/syntasso/kratix-cli/from-api-to-pulumi-pko-program:<version>` (or repo-standard tag strategy)
- Include deterministic env vars in generated pipeline:
  - `PULUMI_COMPONENT_TOKEN`
  - optional traceability vars (for example `PULUMI_SCHEMA_SOURCE`)
- Ensure output placement is consistent with Kratix workflow conventions.
- Set default destination selector strategy for Pulumi preview if this task owns that behavior.

Out of scope:
- Implementing the stage binary internals (that is Task 06).

## File Touchpoints
- `cmd/init_pulumi_component_promise.go`
- helper functions used for pipeline generation (same pattern as operator/crossplane init)
- `test/init_pulumi_component_promise_test.go`

## Implementation Steps
1. Reuse `generateResourceConfigurePipelines(...)` path and adapt args for Pulumi container.
2. Insert Pulumi env vars with stable keys and values.
3. Ensure pipeline emitted correctly in split and flat modes.
4. Add test assertions for container name/image/env.

## Test Plan
Automated:
- Integration tests assert generated workflow contains:
  - correct lifecycle/action placement
  - expected container identity
  - expected env vars
- Run:
```bash
go test ./test/... -run Pulumi
```

Manual inspection:
- Inspect generated `promise.yaml` or split workflow YAML for pipeline correctness.

## Documentation Updates
- Document generated env vars and how users can edit them post-generation (stack/project details handled manually).

## Acceptance Criteria
1. Generated Promise includes resource configure workflow for Pulumi.
2. Workflow references correct container image and env vars.
3. Works consistently in split and non-split generation.
4. Tests guard pipeline contract.

## Definition of Done
- Pulumi init output is workflow-ready and references a concrete runtime stage image.
- Pipeline contract is documented and test-covered.
