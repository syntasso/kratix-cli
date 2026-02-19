# Task 07: Docker Packaging for `component-to-crd`

## Goal
Provide a minimal Docker image that runs the same converter binary and supports stdin/stdout workflows.

## Why This Matters
Planning states binary + Docker delivery, but Docker packaging is currently not implemented.

## Scope
In scope:
- Add a minimal `Dockerfile` under `component-to-crd/`.
- Build static binary in a builder stage and copy into a small runtime image.
- multi-arch publish pipeline for arm and amd so it can run on linux/mac machines
- Set entrypoint to `component-to-crd`.
- Document one smoke test command in `.manual-test/` or README notes.

Out of scope:
- registry push automation
- image signing/SBOM

## Acceptance Criteria
1. `docker build` succeeds from `component-to-crd/`.
2. `docker run` can execute conversion using a mounted schema file.
3. Output from containerized run matches local binary output for same inputs.
4. Existing Go tests remain passing.

## Validation
Automated (local script):
- Add `component-to-crd/.manual-test/10_docker_smoke.sh` that:
  - builds image
  - runs converter on a fixture schema via bind mount
  - compares output with local binary output (`diff`)

Manual:
- Execute script in workspace and confirm parity.

## Definition of Done
- Dockerfile and smoke script committed.
- Containerized invocation documented and validated.
