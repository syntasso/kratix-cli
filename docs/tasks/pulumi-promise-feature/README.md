# Pulumi Promise Feature Task Breakdown

This folder contains self-contained implementation tasks for `docs/pulumi-promise-feature.md`.
Each task is scoped to roughly 1-2 senior-engineer days and includes code, tests, documentation, and usability expectations.

## Recommended Order
1. `task-01-command-scaffold-and-preview-contract.md`
2. `task-02-schema-loader-and-component-selection.md`
3. `task-03-translate-component-inputs-to-crd-spec-schema.md`
4. `task-04-generate-promise-files-split-and-flat.md`
5. `task-05-add-resource-configure-workflow-for-pulumi-program-cr.md`
6. `task-06-build-pulumi-stage-runtime-and-tests.md`
7. `task-07-release-plumbing-and-end-to-end-regression.md`

## Notes
- Tasks 01-05 can be parallelized partially once interfaces are agreed.
- Task 06 depends on workflow contract from Task 05.
- Task 07 should be last to validate full preview readiness.
- Every task can only be completed when all tests pass and the files changed for that task have been locally git committed following semantic conventions.
- Before moving onto another task, be sure to look for refactoring opportunities given the most recent task completion. This could mean within the task solution or across existing commands and the new command. Any refactorings should be added to a new task file with description of what would be done and why it would be valuable.
