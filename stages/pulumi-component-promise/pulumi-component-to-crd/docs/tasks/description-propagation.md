# Task 10: Propagate Pulumi Field Descriptions into Generated CRD Schema

## Goal
Preserve Pulumi schema `description` fields when translating component input properties so generated CRDs include meaningful OpenAPI descriptions.

## Why This Matters
Generated CRDs currently lose field-level documentation from component schemas, reducing usability in kubectl tooling, IDE hover docs, and operator self-service workflows.

## Scope
In scope:
- Translate `description` annotations from Pulumi schema nodes into OpenAPI schema nodes.
- Apply to top-level input properties and nested schemas (object properties, array items, `additionalProperties`, and resolved local refs).
- Ensure CRD emission includes translated descriptions without changing ordering determinism guarantees.
- Document behavior in the design documentation.

Out of scope:
- deriving descriptions from non-standard fields
- markdown normalization or truncation
- adding description generation where source description is missing

## Suggested Implementation
- Update annotation handling in `internal/translate/translate.go` (`applyAnnotations`) to copy string `description` when present.
- Keep existing behavior for `default` and `enum` unchanged.
- Add/extend translator tests in `internal/translate/translate_test.go` to assert descriptions are retained across:
  - primitive field
  - nested object property
  - array item schema
  - local `$ref` with overlay annotations
- Add/extend emission test(s) to verify rendered YAML includes description entries in the expected CRD schema locations.
- Update `docs/pulumi-component-to-crd-design.md` with a short section listing supported annotations (`description`, `default`, `enum`).

## Acceptance Criteria
1. A Pulumi input field with `description` appears as `description` in generated CRD OpenAPI schema.
2. Nested descriptions are preserved for objects, arrays, and map value schemas.
3. Existing `default` and `enum` translation behavior remains unchanged.
4. Translation and emission tests cover the new behavior.
5. Design docs explicitly describe description propagation support.

## Validation
Automated:
- `go test ./...`
- Translation tests assert exact schema maps include expected `description` keys.
- CRD emission tests assert YAML contains expected description lines at `openAPIV3Schema.properties.spec...`.
- review and add to regression tests as needed

## Definition of Done
- Description propagation is implemented with regression tests.
- Documentation reflects supported annotation mapping.
- Existing behavior remains deterministic and backward compatible apart from added description fields.
