# Task 08: Add Kubernetes Structural Validation Gate (Configurable Version)

## Goal
Add a repeatable validation script that checks generated CRDs against a real Kubernetes API server using server-side dry run.

## Why This Matters
Planning requires Kubernetes structural compatibility checks, but no validation gate exists today.

## Scope
In scope:
- Add one workspace-local validation script under `.manual-test/`.
- Script flow:
  1. generate CRD from fixture schema
  2. create a kind cluster for a provided Kubernetes version
  3. run `kubectl apply --dry-run=server -f <generated-crd>`
- Make Kubernetes version configurable (env var or script arg).

Out of scope:
- full CI matrix wiring for all versions
- long-lived cluster lifecycle management

## Acceptance Criteria
1. Script validates CRD structural acceptance on a selected Kubernetes version.
2. Script fails non-zero when server-side dry run fails.
3. Script operates workspace-locally (no `/tmp` artifacts).
4. Converter behavior is unchanged.

## Validation
Automated (script):
- `.manual-test/20_k8s_structural_validate.sh v1.35.x`

Manual follow-up:
- Run same script for additional versions (`v1.32.x` through `v1.35.x`) one-by-one.

## Definition of Done
- Structural validation harness exists and is documented.
- Single-version validation is reproducible and ready for later matrix expansion.
