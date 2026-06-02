# Promise Template

This Promise was generated with:

```bash
kratix init {{ .SubCommand }} {{ .Name }} {{ .ExtraFlags }} --group {{ .Group }} --kind {{ .Kind }}
```

## Updating API properties

To update the Promise API, you can use the `kratix update api` command:

```bash
kratix update api --property name:string --property region- --kind {{ .Kind }}
```

## Updating Workflows

To add workflow containers, you can use the `kratix add container` command:

```bash
kratix add container resource/configure/pipeline0 --image syntasso/postgres-resource:v1.0.0
```

### Building containers

After editing a container's scripts, build and push the image with `kratix build container`:

```bash
# build a specific container
kratix build container resource/configure/pipeline0 --name syntasso-postgres-resource

# build all containers across all workflows
kratix build container --all
```

{{- if eq .SubCommand "pulumi-component-promise" }}
### Pulumi PKO output

The generated Pulumi workflow runs two containers from the same stage codebase:
- `{{ .PulumiGeneratorName }}` emits a PKO `Program`.
- `{{ .PulumiStackGeneratorName }}` emits a PKO `Stack`.

The `Program` container writes deterministic values from existing inputs:
- `program.resources.<component>.type` from `--component` selection.
- `program.resources.<component>.properties` from request `spec`.
- deterministic metadata, namespace and naming from the stage contract.

The `Program` container also reads `PULUMI_SCHEMA_SOURCE` and auto-generates `program.configuration` entries only when values are explicitly trusted in schema config variables (`type`, `default`, `secret`).

The `Stack` container writes deterministic metadata passthrough, `spec.programRef.name`, and `spec.stack` from known request inputs and component identity.

This PKO object generation introduces no additional required environment variables for either the Program or Stack.
If you need additional Pulumi runtime intent, write and add a custom stage container to the Workflow that updates the generated Program or Stack before it is written to stage output.

### Private schema authentication

There are two places where Pulumi code within a Promise may need access to a private registry.
Workflow auth and Stack auth are separate concerns.
The Workflow runs in the cluster where Kratix is running.
The generated `Stack` is reconciled in the scheduled destination cluster, so the referenced Secret must exist in that destination cluster.


### Kratix Workflow runtime

This secret is an environment variable in the Kratix Workflow and can be added or changed in the `{{ .PulumiGeneratorName }}` container:
```yaml
- name: PULUMI_ACCESS_TOKEN
  valueFrom:
    secretKeyRef:
      key: secretKey
      name: secretName
```

To use this Promise, ensure the referenced Secret is present in the namespace where this Workflow runs.
This secret can be populated manually or via the Promise `Dependencies` or `Workflow.Promise.Configure` fields.
Bearer-token schema authentication is only supported for HTTPS schema URLs.

### PKO Stack on Destination

When a request is made to this Promise, a PKO stack will be generated and scheduled to a destination.
On that destination this Stack may need access to a private registry.

This access is set by a separate environment variables in the `{{ .PulumiStackGeneratorName }}` container:

```yaml
- name: PULUMI_STACK_ACCESS_TOKEN_SECRET_NAME
  value: stack
- name: PULUMI_STACK_ACCESS_TOKEN_SECRET_KEY
  value: token
```

For this Stack to work as intended, ensure the referenced Secret is present in the namespace where this Workflow runs.
This secret can be populated manually or via a new container in the `Workflow.Promise.Configure` field if it is common to all resources or the `Workflow.Resource.Configure` if it is unique per request.
{{ end }}


## Updating Dependencies

To add dependencies to the Promise, via a `promise.configure` workflows or the `spec.dependencies` key, you can use the `kratix update dependencies` command:

```bash
kratix update dependencies /dependencies-path
```

## Updating Destination Selectors

To update the Destination Selectors in the Promise Spec, can use the `kratix update destination-selector` flag:

```bash
kratix update destination-selector env=dev
```
