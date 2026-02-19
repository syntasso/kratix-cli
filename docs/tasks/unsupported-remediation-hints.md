# Task 09: Add Remediation Hints for Unsupported Constructs

## Goal
Improve unsupported-construct errors by appending short remediation hints without changing exit code semantics.

## Why This Matters
Planning expects actionable failure messages. Current errors identify construct and path but not suggested next step.

## Scope
In scope:
- Extend `translate.UnsupportedError` to optionally include a hint.
- Add hint mapping for currently rejected keywords/types (for example `oneOf`, `anyOf`, `allOf`, `patternProperties`, `const`).
- Preserve one-line `error: ...` output.
- Keep `exit 3` behavior unchanged.

Out of scope:
- permissive translation fallback
- automatic schema rewrites

## Acceptance Criteria
1. Unsupported errors include component token, path, construct, and a concise hint.
2. Error remains single-line and parseable.
3. Exit code remains `3` for unsupported constructs.
4. Existing non-unsupported errors are unchanged.

## Validation
Automated:
- Update translation and CLI tests to assert hint presence for representative unsupported constructs.

Manual:
- Run converter on a `oneOf` fixture and verify hint text is present.

## Definition of Done
- Unsupported errors are more actionable with no contract regressions.
