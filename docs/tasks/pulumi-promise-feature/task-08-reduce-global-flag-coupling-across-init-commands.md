# Task 08 (Optional): Reduce Global Flag Coupling Across `init` Commands

## Outcome
Refactor `init` command implementations to read flag values from `*cobra.Command` at execution time instead of relying on package-level mutable globals.

## Why This Slice Is Valuable
Current command behavior depends on shared package variables, which increases coupling across commands and makes in-process testing harder. Moving to command-local flag reads improves readability, lowers accidental cross-command interactions, and makes future onboarding safer for junior contributors.

## Scope
In scope:
- `cmd/init*.go` command implementations that currently rely on package globals for command-specific flags.
- Parent `init` persistent flags usage where values are consumed in child commands.
- Integration tests and command-level tests needed to preserve behavior.

Out of scope:
- Functional changes to generated Promise content.
- Renaming user-facing flags or changing CLI contracts.

## Implementation Contract
1. Keep the CLI UX identical:
   - same flag names
   - same required/optional behavior
   - same output and error messages
2. Read values from `cmd.Flags()` / `cmd.InheritedFlags()` (or `cmd.Flag(...)`) inside `RunE`.
3. Prefer a per-command input struct (for readability), for example:
   - `type initPulumiInputs struct { schema string; component string; group string; kind string; ... }`
4. Keep each refactor atomic:
   - one command at a time
   - tests updated in same commit

## Suggested Rollout (Small Slices)
1. Start with `cmd/init_pulumi_component_promise.go` as the pilot.
2. Apply the same pattern to other preview `init` commands:
   - `cmd/init_crossplane_promise.go`
   - `cmd/init_operator_promise.go`
   - `cmd/init_helm_promise.go`
3. Continue to remaining `init` commands:
   - `cmd/init_promise.go`
   - `cmd/init_tf_module_promise.go`
4. Remove now-unused package-level command flag globals after each command is migrated and validated.

## Quality Control Requirements
Automated:
- Run targeted tests after each command migration:
```bash
go test ./cmd/...
go test ./test -ginkgo.focus="init <command-name>"
```
- After all command migrations:
```bash
go test ./cmd/... ./test/...
```

Manual smoke:
- For each migrated command:
  - `--help` works and shows unchanged usage.
  - missing required args/flags return unchanged deterministic errors.
  - valid invocation path behaves exactly as before.

Code review/readability bar:
- Keep diffs focused to one command per commit.
- Avoid mixed behavior changes and refactors in same commit.
- Introduce small helper functions only when they reduce duplication clearly.
- Ensure junior engineer can trace flag source from `RunE` without jumping across multiple files.

## Regression Risk Evaluation
Primary risks:
1. Reading wrong flag set (`Flags` vs `InheritedFlags`) causing empty values at runtime.
2. Changes to error text due to custom validation paths.
3. Subtle ordering issues where defaults differ from previous global variable flow.

Risk level:
- Medium for first migrated command (pattern establishment).
- Low-to-medium for remaining commands once pattern is proven.

Mitigations:
- Preserve Cobra-native validation (`MarkFlagRequired`, arg validators) instead of replacing it.
- Assert exact key error output in integration tests for each command.
- Migrate command-by-command with immediate targeted test runs.

## Acceptance Criteria
1. Migrated commands no longer depend on command-specific package-level mutable globals.
2. CLI user-visible behavior remains unchanged.
3. Tests cover required flag and arg validation paths for each migrated command.
4. Refactor is delivered as small, readable, atomic commits.
