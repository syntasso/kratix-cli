# Task 06: Build `stages/pulumi-promise` Runtime to Emit Program CR

## Outcome
Implement a new stage binary that reads Kratix request input and emits a valid Pulumi Kubernetes Operator Program CR manifest to `/kratix/output`.

## Why This Slice Is Valuable
This completes the execution path so generated workflows can run successfully.

## Scope
In scope:
- Create new stage directory:
  - `stages/pulumi-promise/main.go`
  - `stages/pulumi-promise/Dockerfile`
  - `stages/pulumi-promise/Makefile`
  - `stages/pulumi-promise/CHANGELOG.md`
  - `stages/pulumi-promise/test/...`
- Stage behavior:
  - read request object from Kratix input path
  - map metadata (`name`, `namespace`, labels/annotations as needed)
  - map request `spec` into Program CR payload fields
  - write YAML output file(s) to Kratix output path
- Deterministic naming strategy to avoid collisions.
- Clear errors for malformed input.

Out of scope:
- Provider credential automation.
- Multi-resource expansion.

## Dependencies
Requires Task 05 pipeline contract keys (env var names and expectations).

## File Touchpoints
- `stages/pulumi-promise/*` (new)
- command-side image reference if image/tag path differs
- optional test fixtures under `stages/pulumi-promise/test/assets`

## Implementation Steps
1. Mirror structure from existing stage implementations:
   - `stages/operator-promise/main.go`
   - `stages/crossplane-promise/main.go`
2. Define Program CR API version/kind pinned to tested Pulumi Kubernetes Operator release.
3. Implement mapping logic from request to Program CR.
4. Add unit/integration stage tests for success and failure paths.
5. Ensure Dockerfile builds stage binary and entrypoint matches existing conventions.

## Test Plan
Automated:
- Stage tests:
  - valid input -> expected Program CR
  - missing `spec` -> clear error
  - metadata mapping stability
- Full repo tests for impacted packages.
- Run:
```bash
go test ./stages/pulumi-promise/...
```

Manual smoke:
```bash
cd stages/pulumi-promise && make test
```

## Documentation Updates
- Add stage-specific README section (in main docs or stage docs) describing input/output contract and env vars.

## Acceptance Criteria
1. Stage produces valid Program CR YAML from request input.
2. Output naming and metadata mapping are deterministic.
3. Error paths are explicit and test-covered.
4. Docker build and tests pass for new stage package.

## Definition of Done
- Stage is runnable in Kratix workflow context.
- Tests verify both golden-path and error handling.
- Contract with generated workflow is documented and stable.
