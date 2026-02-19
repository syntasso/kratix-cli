# E2E Example Matrix for `e2e_pulumi_schema_to_crd.sh`

This document lists realistic scenarios for running:

- `component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh`

The script always executes the same high-level flow:

1. Extract schema with `pulumi package get-schema`
2. Run `component-to-crd --in <schema> --component <token>` (with optional CRD identity flags)
3. Validate CRD output

## Prerequisites

- `pulumi` CLI installed and authenticated as needed
- internet access for dependency/plugin downloads and registry-backed examples
- from repo root, build the test binary once:

```bash
component-to-crd/.manual-test/00_build_binary.sh
```

Quick default run (no arguments):

```bash
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh
```

## Docker Packaging and Tests

Build only for local architecture:

```bash
IMAGE_TAG=component-to-crd:local component-to-crd/scripts/docker_build_local.sh
```

Build and push a multi-arch image (Linux amd64 + arm64):

```bash
IMAGE_TAG=ghcr.io/<org>/component-to-crd:<tag> \
  component-to-crd/scripts/docker_buildx_push_multiarch.sh
```

Run a containerized conversion test with an input component token:

```bash
component-to-crd/scripts/docker_build_local.sh

docker run --rm \
  --mount "type=bind,src=$PWD/component-to-crd/.manual-test,dst=/work,readonly" \
  component-to-crd:local \
  --in pulumi/tests/integration/component_provider/nodejs/component-provider-host/provider
```

Expected result:
- command exits `0`
- CRD YAML is printed to stdout
- output includes identity for `pkg:index:Thing` (for example `kind: Thing`)

## Task 03 URL Input Manual Checks

Task 03 added URL support for `--in` (local file paths still work unchanged).

Run the dedicated live URL script:

```bash
component-to-crd/.manual-test/08_url_input_live_registry.sh
```

This script validates:
- `--in https://www.pulumi.com/registry/packages/eks/schema.json --component eks:index:Cluster`
  - expected: URL fetch succeeds, then current preflight behavior returns exit `2` for unsupported `#/resources/...` refs.
- `--in https://www.pulumi.com/registry/packages/eks/does-not-exist.json --component eks:index:Cluster`
  - expected: exit `2` with `error: fetch input schema URL: unexpected status 404 for ...`.

To include these live URL checks in the full manual suite:

```bash
RUN_INTERNET_TESTS=1 component-to-crd/.manual-test/99_run_all.sh
```

## Task 04 CRD Identity Manual Checks

Task 04 adds optional CRD identity flags:
- `--group`
- `--version`
- `--kind`
- `--plural`
- `--singular`

Validation notes:
- `--kind` must match `^[A-Za-z][A-Za-z0-9]*$` (for example `ServiceDeployment`).
- `--version`, `--plural`, and `--singular` must be DNS-label-like (lowercase alphanumeric with optional internal `-`).
- `--group` must be DNS-subdomain-like.

Validate default identity values:

```bash
component-to-crd/.manual-test/component-to-crd \
  --in component-to-crd/.manual-test/schema.valid.json
```

Expected identity snippets:
- `metadata.name: "components.components.pulumi.local"`
- `spec.group: components.pulumi.local`
- `spec.names.kind: Component`
- `spec.names.plural: components`
- `spec.names.singular: component`
- `spec.versions[0].name: v1alpha1`

Validate custom identity values:

```bash
component-to-crd/.manual-test/component-to-crd \
  --in component-to-crd/.manual-test/schema.valid.json \
  --group apps.example.com \
  --version v1 \
  --kind ServiceDeployment \
  --plural servicedeployments \
  --singular servicedeployment
```

Expected identity snippets:
- `metadata.name: "servicedeployments.apps.example.com"`
- `spec.group: apps.example.com`
- `spec.names.kind: ServiceDeployment`
- `spec.names.plural: servicedeployments`
- `spec.names.singular: servicedeployment`
- `spec.versions[0].name: v1`

Validate invalid identity handling (`exit 2`, single-line `error:`):

```bash
component-to-crd/.manual-test/component-to-crd \
  --in component-to-crd/.manual-test/schema.valid.json \
  --group bad_group
```

## Example 1: Simple Local Component (previous simple case)

Source:
- `pulumi/tests/integration/namespaced_component`

Command:

```bash
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh \
  --component namespaced-component:index:MyComponent \
  --schema-source pulumi/tests/integration/namespaced_component \
  --package-name namespaced-component \
  --expect-crd-contains anInput:
```

## Example 2: Complex Local Component (current local richer case)

Source:
- `pulumi/tests/integration/component_provider/nodejs/component-provider-host/provider`

Command:

```bash
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh \
  --component nodejs-component-provider:index:MyComponent \
  --schema-source pulumi/tests/integration/component_provider/nodejs/component-provider-host/provider \
  --package-name nodejs-component-provider \
  --install-plugin resource:random:v4.18.0 \
  --expect-schema-contains '"aComplexTypeInput"' \
  --expect-schema-contains '"enumInput"' \
  --expect-crd-contains aComplexTypeInput: \
  --expect-crd-contains nestedComplexType: \
  --expect-crd-contains enum:
```

## Example 3: Internet Package - Pulumi EKS

Registry page:
- https://www.pulumi.com/registry/packages/eks/

Schema URL:
- https://www.pulumi.com/registry/packages/eks/schema.json

Command:

```bash
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh \
  --component eks:index:Cluster \
  --schema-source eks@4.2.0 \
  --package-name eks
```

## Example 4: Internet Package - Kubernetes Cert Manager

Registry page:
- https://www.pulumi.com/registry/packages/kubernetes-cert-manager/

Schema URL:
- https://www.pulumi.com/registry/packages/kubernetes-cert-manager/schema.json

Command:

```bash
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh \
  --component kubernetes-cert-manager:index:CertManager \
  --schema-source kubernetes-cert-manager \
  --package-name kubernetes-cert-manager
```

## Example 5: Internet Package - Kubernetes Ingress NGINX

Registry page:
- https://www.pulumi.com/registry/packages/kubernetes-ingress-nginx/

Schema URL:
- https://www.pulumi.com/registry/packages/kubernetes-ingress-nginx/schema.json

Command:

```bash
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh \
  --component kubernetes-ingress-nginx:index:IngressController \
  --schema-source kubernetes-ingress-nginx \
  --package-name kubernetes-ingress-nginx
```

## Running All Five Examples

Run each command in order. This keeps logs and artifacts isolated per example under:

- `component-to-crd/.manual-test/work.e2e.<work-name>/`

```bash
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh \
  --component namespaced-component:index:MyComponent \
  --schema-source pulumi/tests/integration/namespaced_component \
  --package-name namespaced-component \
  --expect-crd-contains anInput:

component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh \
  --component nodejs-component-provider:index:MyComponent \
  --schema-source pulumi/tests/integration/component_provider/nodejs/component-provider-host/provider \
  --package-name nodejs-component-provider \
  --install-plugin resource:random:v4.18.0 \
  --expect-schema-contains '"aComplexTypeInput"' \
  --expect-schema-contains '"enumInput"' \
  --expect-crd-contains aComplexTypeInput: \
  --expect-crd-contains nestedComplexType: \
  --expect-crd-contains enum:

component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh \
  --component eks:index:Cluster \
  --schema-source eks@4.2.0 \
  --package-name eks

component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh \
  --component kubernetes-cert-manager:index:CertManager \
  --schema-source kubernetes-cert-manager \
  --package-name kubernetes-cert-manager

component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh \
  --component kubernetes-ingress-nginx:index:IngressController \
  --schema-source kubernetes-ingress-nginx \
  --package-name kubernetes-ingress-nginx
```

Tip for internet-backed examples:
- If a component token is wrong for a package/version, extract schema then list component tokens:

```bash
pulumi package get-schema eks@4.2.0 > component-to-crd/.manual-test/work.e2e.eks-token-discovery.schema.json
python3 - <<'PY'
import json
with open('component-to-crd/.manual-test/work.e2e.eks-token-discovery.schema.json', 'r', encoding='utf-8') as f:
    schema = json.load(f)
for token, resource in sorted(schema.get('resources', {}).items()):
    if resource.get('isComponent') is True:
        print(token)
PY
```
