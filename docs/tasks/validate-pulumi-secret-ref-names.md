# Task: Validate Pulumi Secret Reference Names for Kubernetes Compatibility

## Outcome
Block `init pulumi-component-promise` when either `--schema-bearer-token-secret` or `--stack-access-token-secret` includes a Secret name that is not valid for a Kubernetes Secret reference.

## Why This Slice Is Valuable
The new Pulumi auth flags currently validate only the `SECRET_NAME:KEY` shape.
That means a user can generate a Promise that looks valid at init time but later fails when Kubernetes rejects the referenced Secret name.

Adding early validation keeps the CLI feedback loop short, makes the generated output more trustworthy, and is easier for junior engineers to reason about than deferring the failure to runtime.

## Scope
In scope:
- Validation for the Secret name portion of:
  - `--schema-bearer-token-secret`
  - `--stack-access-token-secret`
- Clear user-facing CLI errors that stop Promise generation before files are written.
- Unit and integration tests that cover valid and invalid cases.

Out of scope:
- Validation of the Secret key portion.
- Changes to generated YAML for valid inputs.
- Broader refactors of init flag parsing.

## Implementation Contract
1. Keep the existing `SECRET_NAME:KEY` parsing contract.
2. After parsing, validate the Secret name against Kubernetes Secret naming rules used for object references.
3. Treat the Secret name as invalid if it is not a valid DNS-1123 subdomain.
4. Reject invalid input before any Promise files are written.
5. Return a clear error message that includes:
   - which flag failed
   - the invalid Secret name
   - that the name must be a valid Kubernetes Secret name

Example error shape:

```text
parse --schema-bearer-token-secret: secret name "Pulumi_Secret" is not a valid Kubernetes Secret name
```

Implementation note:
- Prefer reusing Kubernetes validation helpers rather than duplicating the naming rules in custom string logic.
- Keep the validation path close to the existing secret flag parsing so a reader can trace parsing and validation in one place.

## Suggested Approach
1. Parse `SECRET_NAME:KEY` exactly as today.
2. Add a small helper that validates the parsed Secret name.
3. Reuse that helper for both Pulumi auth flags.
4. Return early from `initPulumiComponentPromiseFromSelection(...)` when validation fails.

## Test Plan
Automated:

```bash
go test ./cmd/...
go test ./test/... -ginkgo.focus="init pulumi-component-promise"
```

Add or update tests for:
- valid `--schema-bearer-token-secret` value still succeeds
- valid `--stack-access-token-secret` value still succeeds
- invalid schema Secret name fails with a clear deterministic error
- invalid Stack Secret name fails with a clear deterministic error
- no files are created on those failure paths

Recommended invalid examples:
- `PulumiSecret:accessToken`
- `pulumi_secret:accessToken`
- `pulumi.secret-:accessToken`

Recommended valid examples:
- `pulumi-schema-auth:accessToken`
- `pulumi-api-secret:accessToken`

## Regression Risk Evaluation
Primary risks:
1. Using validation rules that are stricter or looser than Kubernetes object naming.
2. Changing existing error text in a way that makes tests brittle without improving clarity.
3. Validating too late, after files have already been written.

Mitigations:
- Use Kubernetes-provided validation helpers.
- Add focused unit tests for the helper and integration tests for CLI behaviour.
- Keep validation before any call that writes output files.

## Acceptance Criteria
1. `init pulumi-component-promise` rejects invalid Secret names for both new Pulumi auth flags.
2. Error messages clearly identify the offending flag and invalid name.
3. No Promise files are written when validation fails.
4. Existing behaviour for valid values remains unchanged.
