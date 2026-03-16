# Pulumi schema fetch and private registry authentication research

## Context

This note captures how `kratix init pulumi-component-promise` currently fetches schema data and what would be required to support private schema URLs, including OIDC-oriented approaches and Pulumi Cloud auth for generated PKO `Stack` objects.

## How schema fetch works today

### During `init` (CLI machine)

1. `init pulumi-component-promise` reads `--schema` and calls `pulumi.LoadSchema(...)`.
2. If `--schema` is an `http` or `https` URL, `internal/pulumi/schema_loader.go` fetches it with a plain `client.Get(rawURL)`.
3. No authentication headers, no custom transport, and no credentials callback are used.
4. Non-`200` responses fail command execution.

### During workflow runtime (cluster container)

1. Promise generation stores schema source in workflow env var `PULUMI_SCHEMA_SOURCE` for the `pulumi-program-generator` container.
2. The stage entrypoint reads `PULUMI_SCHEMA_SOURCE` and calls the same `pulumi.LoadSchema(...)` path.
3. The runtime container therefore repeats the same unauthenticated HTTP GET behaviour.

### During PKO stack reconciliation

1. The generated `pulumi-stack-generator` stage emits a PKO `Stack` custom resource.
2. Today that generated `Stack` contains deterministic metadata plus `spec.programRef.name` and `spec.stack`.
3. It does not currently add Pulumi Cloud auth such as `spec.envRefs.PULUMI_ACCESS_TOKEN`.

## Current implications

- Public HTTP(S) URLs work.
- Local paths trigger warnings at `init` because the runtime container cannot read the developer workstation filesystem.
- Private URLs only work if reachable without extra auth logic (for example, pre-signed URL or credentials embedded in URL, which is usually not desirable).

## Can OIDC be used?

Short answer: yes in some environments, but not directly for all registries.

### What OIDC gives a Kubernetes workload

Kubernetes can project a service account token into a pod with configurable `audience` and `expirationSeconds`.

- Source: https://kubernetes.io/docs/concepts/storage/projected-volumes/

That token can be used directly only when the target service trusts the cluster OIDC issuer, or indirectly by exchanging it for another access token.

### Cloud workload identity (token exchange)

Cloud platforms commonly exchange Kubernetes service account tokens for short-lived access tokens:

- AWS IRSA: projected token can call `AssumeRoleWithWebIdentity`.
  - Source: https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
- GKE Workload Identity Federation: Kubernetes token exchanged through Google Security Token Service.
  - Source: https://cloud.google.com/kubernetes-engine/docs/concepts/workload-identity
- Azure workload identity federation: projected service account token exchanged for Microsoft Entra token.
  - Source: https://learn.microsoft.com/en-us/azure/azure-arc/kubernetes/conceptual-workload-identity

### GitHub-specific reality

GitHub private content access is token-based (PAT or GitHub App installation token), not direct trust of arbitrary Kubernetes OIDC workload tokens.

- Auth methods for REST API are token-centric.
  - Source: https://docs.github.com/en/rest/authentication/authenticating-to-the-rest-api
- GitHub App installation auth uses short-lived installation tokens.
  - Source: https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation
- Repository contents endpoint expects token auth for private content.
  - Source: https://docs.github.com/en/rest/repos/contents

Inference: for GitHub-hosted private schema files, OIDC can still be part of the flow, but usually through a broker/exchange step that mints a GitHub token, not by sending the Kubernetes OIDC token directly to GitHub.

## Easiest viable configuration direction in generated Promise

The simplest broadly-compatible model is:

1. Keep `PULUMI_SCHEMA_SOURCE` as URL.
2. Add optional auth token input for HTTP `Authorization: Bearer ...`.
3. Source that token from Kubernetes `Secret` or from a token file written by workload-identity exchange logic.
4. Separately, add optional PKO `Stack.spec.envRefs.PULUMI_ACCESS_TOKEN` so the reconciled stack can authenticate to Pulumi Cloud.

This keeps runtime deterministic and avoids provider-specific logic in the loader.

## Pulumi Cloud access shape for generated `Stack` objects

Pulumi documents Stack-level cloud access through `spec.envRefs`, for example:

```yaml
envRefs:
  PULUMI_ACCESS_TOKEN:
    type: Secret
    secret:
      name: pulumi-api-secret
      key: accessToken
```

Source: https://www.pulumi.com/docs/iac/guides/continuous-delivery/pulumi-kubernetes-operator/#configure-pulumi-cloud-access

Inference: workflow-container schema auth and PKO Stack auth should be treated as two related but distinct configuration paths. One authenticates `pulumi.LoadSchema(...)`; the other authenticates the Pulumi Kubernetes Operator when it reconciles the generated `Stack`.

## In-scope constraints observed in repository

- Stage image is distroless and contains only the two compiled binaries.
- Schema loading logic is shared by CLI init and stage runtime.
- Existing tests already cover URL loading, status handling, and runtime env contract.

These constraints favour adding a small, generic auth hook in `internal/pulumi/schema_loader.go` instead of embedding cloud-specific SDKs.
