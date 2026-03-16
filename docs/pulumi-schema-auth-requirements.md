# Ordered requirements for private Pulumi schema authentication

## Goal

Allow generated Pulumi Promise workflows to fetch private remote schema URLs safely, and allow generated PKO `Stack` objects to authenticate to Pulumi Cloud, with the smallest possible sequence of changes.

## Recommendation summary

Implement a generic HTTP auth extension that can work generically for any private registry. Future changes can introduce OIDC or other authentication methods.

## Requirement 1: Add optional bearer token support in schema loader

### Change

- Extend URL fetch logic to send `Authorization: Bearer <token>` when configured.
- Keep default behaviour unchanged when no auth is provided.
- Prefer explicit env vars used by both CLI and stage runtime:
  - `PULUMI_ACCESS_TOKEN`
  - `PULUMI_ACCESS_TOKEN_FILE` (file path wins only if token env is empty)

### Why first

- Minimal code change.
- Unblocks private schema endpoints immediately.
- Works with GitHub tokens, registry tokens, and exchanged workload-identity tokens.

### Easy verification

- Unit test that header is absent by default.
- Unit test that header is present when env var is set.
- Unit test that token file path is read and used.
- Existing URL/status tests remain green when bearer token is not provided.

### UX impact

- New optional env vars only.
- Existing users see no change.

## Requirement 2: Wire generated Promise workflow to secret-backed token (opt-in)

### Change

- Add optional init flag to scaffold secret reference into generated `pulumi-program-generator` container, for example:
  - `--schema-bearer-token-secret SECRET_NAME:KEY`
- Generated workflow should set `PULUMI_ACCESS_TOKEN` via `valueFrom.secretKeyRef`.

### Why second

- Keeps auth ergonomic in generated Promise without forcing manual YAML edits.
- Still provider-agnostic.

### Easy verification

- Integration test: generated workflow includes `secretKeyRef` only when flag is used.
- Regression test: generated output unchanged when flag is omitted.

### UX impact

- New optional CLI flag for secure auth setup.
- README template needs one short section showing secret creation and flag usage.
- Include in README how to set this flag manually if not done via the init flag.

## Requirement 3: Add Pulumi Cloud auth to generated `Stack` objects

### Change

- Add optional init flag to scaffold Pulumi Kubernetes Operator `envRefs` into generated `Stack` output.
- Generated `Stack` should set `spec.envRefs.PULUMI_ACCESS_TOKEN` using the PKO secret reference shape documented by Pulumi:

```yaml
envRefs:
  PULUMI_ACCESS_TOKEN:
    type: Secret
    secret:
      name: pulumi-api-secret
      key: accessToken
```
- Make sure the values are set via env var in the body of the Promise so that they are easy to update and access.

### Why third

- Schema fetch auth only helps the workflow read the component schema.
- The generated `Stack` still needs its own Pulumi Cloud credentials when PKO reconciles it.
- Reuses the same secret-backed, opt-in model instead of introducing a different auth mechanism.

### Easy verification

- Integration test: generated `Stack` includes `spec.envRefs.PULUMI_ACCESS_TOKEN` only when the flag is used.
- Regression test: generated `Stack` output stays unchanged when the flag is omitted.

### UX impact

- New optional CLI flag, or an extension of the existing secret flag if the same secret should drive both schema fetch and Stack auth.
- README template needs a short note explaining that workflow auth and Stack auth are separate concerns. Clearly stating that the Workflow is run in the cluster where Kratix is running while the stack is running in the scheduled destination cluster and the secret must exist in that cluster.
- Include the Pulumi-documented `envRefs` example so users can patch generated output manually if needed.

## Suggested implementation order

1. Requirement 1
2. Requirement 2
3. Requirement 3

This sequence keeps each increment small, testable, and independently valuable.
