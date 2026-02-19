# Task: Skip Untranslatable Paths and Log During CRD Generation

## Problem
Selected component graphs in real schemas can include constructs that cannot be translated into valid CRD fields (for example composition keywords, shapes, or references not supported by current translation rules).

Current behavior can fail the entire run (`exit 3`) when one path is unsupported, even if most of the component can still be translated.

## Goal
Make translation resilient by skipping only the untranslatable schema paths, continuing CRD generation for the rest of the selected component, and logging skipped paths in a deterministic and parseable way.

## Desired UX State
- CLI generates CRD YAML when at least one translatable field remains.
- Any skipped field path is logged with component token, path, and reason.
- Logs are deterministic for equivalent inputs.
- Generated YAML remains deterministic and stable.

## Non-goals
- No attempt to fully implement every unsupported schema construct.
- No remote schema dereferencing implementation.
- No change to output destination contract (CRD YAML still on `stdout`).

## Functional Requirements
1. During translation, if a path cannot be translated into a CRD field, skip that path instead of failing the whole command.
2. Emit a log entry for each skipped path containing:
   - selected component token
   - schema path
   - concise reason
3. Keep skipped-path logging deterministic (sorted/stable order).
4. Exclude skipped properties from emitted OpenAPI schema.
5. Required-field handling:
   - if a required property is skipped, remove it from emitted `required` to keep schema valid
   - if this results in empty `required`, omit `required` for that object as normal
6. If translation results in zero translatable top-level `spec` properties, fail with a user/input error (`exit 2`) and actionable message.
7. CLI help text (`--help`) must document:
   - that untranslatable field paths are skipped (not hard-failed) in this mode
   - where skip details are reported (`stderr`)
   - the currently known classes of untranslatable constructs that may be skipped (for example composition keywords like `oneOf`/`anyOf`/`allOf`, unsupported schema keywords, unresolved refs outside supported handling)

## Exit/Error Contract
- Keep existing exit code contract:
  - `0` success
  - `2` user/input/preflight errors (including no translatable `spec` fields)
  - `3` reserved for hard unsupported conditions that are still intentionally non-skippable
  - `4` output/serialization/write failures
- Keep user-facing errors single-line parseable (`error: ...`).
- Keep skip logs single-line parseable (for example `warn: component=... path=... reason=...`).

## Proposed Implementation
1. Introduce a translation issue collector that records skippable path failures.
2. Update translation traversal to:
   - attempt translation per property/node
   - on skippable failures, record issue and continue
   - on non-skippable failures, preserve current hard-fail behavior
3. Thread warning output through CLI orchestration and print deterministic skip logs to `stderr`.
4. Normalize `required` arrays after skip filtering.
5. Add a final guard for empty translated `spec` fields.

## Tests (Required)
1. `internal/translate/translate_test.go`
   - unsupported field path is skipped and logged via collector
   - required list is adjusted when skipped fields were required
   - deterministic translated output and deterministic skip ordering
   - hard non-skippable errors still fail
2. `cmd/component-to-crd/main_test.go`
   - regression: schema with mixed translatable/untranslatable fields succeeds and logs skipped paths
   - regression: all top-level fields untranslatable fails with `exit 2`
   - deterministic stderr log ordering assertion for skipped paths
   - `--help` output includes untranslatable-field guidance and examples of construct classes
3. Existing tests remain green (`go test ./...`).

## Acceptance Criteria
- Real schemas with partial unsupported paths produce usable CRD output instead of failing wholesale.
- Skipped paths are visible and actionable in deterministic logs.
- Output schema remains structurally valid (including `required` coherence).
- Runs still fail clearly when nothing translatable remains at top-level `spec`.
- All tests pass.

## Verification
Run from `component-to-crd`:

```bash
go test ./...
```

Manual validation:

```bash
go run ./cmd/component-to-crd \
  --in https://raw.githubusercontent.com/pulumi/pulumi-eks/master/provider/cmd/pulumi-resource-eks/schema.json \
  --component eks:index:Cluster > component-to-crd/.manual-test/work.skip/cluster.crd.yaml
```

Confirm:
- command exits `0` when at least one top-level `spec` field is translatable
- `stderr` includes parseable skip logs for ignored paths
- generated YAML excludes skipped fields
- if forced to a schema where all top-level `spec` fields are untranslatable, command exits `2`
