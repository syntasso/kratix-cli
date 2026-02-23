# AGENTS.md

This file gives guidance for LLM agents contributing to this repository.

## Intent
- Keep this codebase junior-friendly.
- Follow XP-style delivery: small, safe, test-backed changes.
- Prefer clarity and maintainability over cleverness.

## Working Agreement
- Make the smallest useful change first.
- Explain intent in code and commit messages so a junior engineer can follow the reasoning.
- Keep behavior deterministic where existing code expects it (especially in stage tooling and generated output).
- When you change behavior, update tests in the same change.
- When a test fails, start by confirming the code isn't broken before changing any tests.
- When relevant, update docs in the same change.
- This codebase works in British English for all documentation.

## XP Principles to Apply Here
- Work in tiny slices.
- Use test-first when practical:
  - add or update a failing test
  - make it pass with minimal code
  - refactor safely
- Prefer simple design and explicit contracts.
- Codify all common tasks as repeatable and deterministic Make targets.
- Do not batch unrelated refactors with behavior changes.

## Commit Discipline (Important)
- Use atomic commits: one logical change per commit.
- Every completed code change must be captured in a local commit before handoff.
- A commit should be understandable on its own by a junior engineer.
- Keep commits mixed only when needed for coherence:
  - code + tests for that exact behavior
  - docs for that exact behavior
- Avoid “drive-by” edits in unrelated files.

## Repo-Specific Guardrails
- All tasks should be codified as Make targets, for example:
  - `make build`
  - `make test`
- For Go-focused areas, run targeted `go test` for touched packages, then broaden scope as needed.
- Preserve existing command/output contracts when touching CLI behavior.
- If you change that adapter’s user-visible behavior, update that design doc and tests together.
- Keep contributor workflow/process guidance in this `AGENTS.md`, not in the Pulumi design ADR.

## Definition of Done for Agent Changes
- Change is small and focused.
- Tests cover the changed behavior.
- Docs are updated when behavior/contracts changed.
- Diff and commit history are easy for a junior engineer to review end-to-end.
