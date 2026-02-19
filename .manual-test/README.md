# Manual E2E Guide

This directory contains manual end-to-end test workflows for `component-to-crd`.

## What To Run

Host-binary conversion path:

```bash
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh
```

Docker conversion path (no host binary invocation):

```bash
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd_docker.sh
```

Direct `docker run` example:

```bash
IMAGE_TAG=component-to-crd:local component-to-crd/scripts/docker_build_local.sh

docker run --rm \
  component-to-crd:local \
  --in https://www.pulumi.com/registry/packages/eks/schema.json \
  --component eks:index:Cluster
```

Both scripts follow the same flow:
1. Extract Pulumi schema with `pulumi package get-schema`.
2. Convert schema to CRD YAML for a selected component.
3. Validate expected output and write artifacts under `component-to-crd/.manual-test/work/`.

## Prerequisites

Required:
- `pulumi` CLI

For Docker path:
- `docker`

For internet-backed examples:
- network access for schema/plugin downloads

## Fast Checks

Run unit + CLI + regression tests from the package root:

```bash
cd component-to-crd && go test ./...
```

Run internet-backed regression subset:

```bash
cd component-to-crd && RUN_INTERNET_TESTS=1 go test ./regression-test -run URLInputLiveRegistry
```

## Script Usage

Show options:

```bash
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh --help
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd_docker.sh --help
```

Common flags (both scripts):
- `--component <token>`
- `--schema-source <source>`
- `--work-name <name>`
- `--package-name <name>`
- `--install-plugin <kind:name:version>` (repeatable)
- `--expect-schema-contains <text>` (repeatable)
- `--expect-crd-contains <text>` (repeatable)
- `--skip-install`

Docker-only flags:
- `--image-tag <tag>`
- `--skip-image-build`

## Docker Image Commands

Build local image:

```bash
IMAGE_TAG=component-to-crd:local component-to-crd/scripts/docker_build_local.sh
```

Build/push multi-arch image:

```bash
IMAGE_TAG=ghcr.io/<org>/component-to-crd:<tag> \
  component-to-crd/scripts/docker_buildx_push_multiarch.sh
```

## CRD Identity Sanity Checks

Build host binary:

```bash
component-to-crd/scripts/build_binary
```

Default identity:

```bash
component-to-crd/bin/component-to-crd \
  --in component-to-crd/regression-test/testdata/schemas/schema.valid.json
```

Custom identity:

```bash
component-to-crd/bin/component-to-crd \
  --in component-to-crd/regression-test/testdata/schemas/schema.valid.json \
  --group apps.example.com \
  --version v1 \
  --kind ServiceDeployment \
  --plural servicedeployments \
  --singular servicedeployment
```

Invalid identity (`exit 2`, single-line `error:`):

```bash
component-to-crd/bin/component-to-crd \
  --in component-to-crd/regression-test/testdata/schemas/schema.valid.json \
  --group bad_group
```

## Example Matrix

Example 1: simple local component

```bash
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh \
  --component namespaced-component:index:MyComponent \
  --schema-source pulumi/tests/integration/namespaced_component \
  --package-name namespaced-component \
  --expect-crd-contains anInput:
```

Example 2: complex local component

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

Example 3: Pulumi EKS from internet

```bash
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh \
  --component eks:index:Cluster \
  --schema-source eks@4.2.0 \
  --package-name eks
```

Example 4: Kubernetes Cert Manager from internet

```bash
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh \
  --component kubernetes-cert-manager:index:CertManager \
  --schema-source kubernetes-cert-manager \
  --package-name kubernetes-cert-manager
```

Example 5: Kubernetes Ingress NGINX from internet

```bash
component-to-crd/.manual-test/e2e_pulumi_schema_to_crd.sh \
  --component kubernetes-ingress-nginx:index:IngressController \
  --schema-source kubernetes-ingress-nginx \
  --package-name kubernetes-ingress-nginx
```

Tip for internet-backed examples: if a token is wrong for a package/version, extract schema and list component tokens.

```bash
pulumi package get-schema eks@4.2.0 > component-to-crd/.manual-test/work/e2e.eks-token-discovery.schema.json
python3 - <<'PY'
import json
with open('component-to-crd/.manual-test/work/e2e.eks-token-discovery.schema.json', 'r', encoding='utf-8') as f:
    schema = json.load(f)
for token, resource in sorted(schema.get('resources', {}).items()):
    if resource.get('isComponent') is True:
        print(token)
PY
```
