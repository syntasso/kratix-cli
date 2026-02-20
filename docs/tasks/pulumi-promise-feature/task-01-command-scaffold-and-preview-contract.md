# Task 01: Add `init pulumi-component-promise` Command Scaffold and Preview Contract

## Outcome
Introduce a new preview CLI command that parses and validates Pulumi-specific inputs, but does not yet need full schema translation or stage output wiring.

## Why This Slice Is Valuable
This creates the stable command surface (`help`, flags, error handling, preview warning, usage examples) so follow-on work can implement internals behind a locked user contract.

## Scope
In scope:
- Register `kratix init pulumi-component-promise` under `init` command tree.
- Add Pulumi-specific flags:
  - `--schema` (required)
  - `--component` (optional)
- Reuse existing global init flags already provided by parent init command:
  - `--group`, `--kind`, `--version`, `--plural`, `--dir`, `--split`
- Mark command as preview in `Short`/`Long` text.
- Call `printPreviewWarning()` at runtime.
- Implement argument/flag validation with deterministic user-facing errors.

Out of scope:
- Translating schema into CRD.
- Writing promise files.
- Generating workflows.

## Inputs and UX Contract
Expected command shape:
```bash
kratix init pulumi-component-promise PROMISE-NAME --schema PATH_OR_URL [--component TOKEN] [--group GROUP] [--kind KIND] [--version VERSION] [--plural PLURAL] [--split] [--dir DIR]
```

Validation behavior to implement now:
- Missing promise name: fail with usage + clear error.
- Missing `--schema`: fail with clear error.
- Extra positional args: fail.
- `--help`: success, no preview warning printed.

## File Touchpoints
- `cmd/init.go` (command wiring)
- `cmd/init_pulumi_component_promise.go` (new file)
- `cmd/preview_warning.go` (reuse only)
- `test/init_pulumi_component_promise_test.go` (new integration-style command tests)

## Implementation Steps
1. Add new Cobra command file following patterns in:
   - `cmd/init_helm_promise.go`
   - `cmd/init_crossplane_promise.go`
2. Define constants for command name and examples.
3. Add minimal `RunE` that:
   - validates args/flags
   - prints preview warning
   - returns `nil` for now (or TODO path marker)
4. Register command from `init()`.
5. Add examples for local schema and URL schema.

## Test Plan
Automated:
- Add tests in `test/init_pulumi_component_promise_test.go` covering:
  - command exists under `kratix init --help`
  - `--help` output includes preview language and examples
  - missing name fails
  - missing `--schema` fails
  - valid minimal invocation reaches `RunE` success path
- Run:
```bash
go test ./cmd/... ./test/...
```

Manual smoke:
```bash
go run ./cmd/kratix/main.go init pulumi-component-promise mypromise --schema ./testdata/schema.json --help
```

## Documentation Updates
- Add command mention in root `README.md` init section as preview command (single short paragraph is enough for this task).

## Acceptance Criteria
1. Command is discoverable in CLI help and wired under `kratix init`.
2. Preview warning is printed on real execution paths.
3. Required arguments and `--schema` validation are enforced with deterministic errors.
4. Integration tests cover success + failure validation paths.
5. README mentions command availability as preview.

## Definition of Done
- Code merged with tests passing.
- User can run the command and understand required inputs from help text alone.
- README includes a usable command example.
