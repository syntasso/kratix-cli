# Task 04: Generate Promise Files (Flat and `--split`) from Translated Schema

## Outcome
Wire translated schema into standard Kratix Promise file generation, producing complete bootstrap artifacts users can run.

## Why This Slice Is Valuable
This turns parsing + translation into tangible output users can adopt immediately.

## Scope
In scope:
- Generate standard artifacts for both output modes:
  - flat: `promise.yaml`, `example-resource.yaml`, `README.md`
  - split: `api.yaml`, workflow files, dependencies where applicable, `example-resource.yaml`, `README.md`
- Use translated Pulumi schema as Promise API source.
- Preserve existing generation conventions used by other init commands.
- Include command reconstruction line in generated README.

Out of scope:
- Full stage implementation details.
- Release automation.

## Suggested Reuse
Use existing helper paths from current init flows:
- `getFilesToWrite(...)`
- `writePromiseFiles(...)`
- template files under `cmd/templates/promise/`

## File Touchpoints
- `cmd/init_pulumi_component_promise.go`
- `cmd/templates/promise/README.md.tpl` (only if reconstruction text needs extension)
- `test/init_pulumi_component_promise_test.go`
- fixtures under `test/assets/` for expected output snapshots

## Implementation Steps
1. Build Promise object from group/kind/version/plural defaults + translated schema.
2. Route through existing split/non-split write flow.
3. Generate consistent example resource from translated required/optional fields.
4. Ensure generated README includes exact init command, including schema and component args used.
5. Add deterministic fixture expectations for both modes.

## Test Plan
Automated:
- Integration tests validating:
  - non-split file set and content
  - split file set and content
  - README reconstruction command
  - expected schema fields appear in API
- Run:
```bash
go test ./test/... -run Pulumi
```

Manual smoke:
```bash
kratix init pulumi-component-promise mypromise --schema ./test/assets/pulumi/schema.valid.json
kratix init pulumi-component-promise mypromise --schema ./test/assets/pulumi/schema.valid.json --split
```

## Documentation Updates
- Update root `README.md` with examples for both split and non-split output.

## Acceptance Criteria
1. Command writes complete usable Promise scaffolding in both modes.
2. Generated API includes translated Pulumi spec schema.
3. README reconstruction command is accurate and runnable.
4. Integration tests validate file content and structure for both modes.

## Definition of Done
- Users can run command and get valid Promise assets without manual file authoring.
- Tests cover both generation modes end-to-end.
- Documentation shows practical usage examples.
