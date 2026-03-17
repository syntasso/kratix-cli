# Promise Template

This Promise was generated with:

```
kratix init {{ .SubCommand }} {{ .Name }} {{ .ExtraFlags }} --group {{ .Group }} --kind {{ .Kind }}
```

## Updating API properties

To update the Promise API, you can use the `kratix update api` command:

```
kratix update api --property name:string --property region- --kind {{ .Kind }}
```

## Updating Workflows

To add workflow containers, you can use the `kratix add container` command:

```
kratix add container resource/configure/pipeline0 --image syntasso/postgres-resource:v1.0.0
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

To fetch a private remote schema, create a Secret in the namespace where this Workflow runs and set `PULUMI_ACCESS_TOKEN` on `{{ .PulumiGeneratorName }}`.

Example:

```bash
kubectl create secret generic pulumi-schema-auth --from-literal=accessToken='<token>'
```

{{- if .SchemaBearerTokenSecret }}

This Promise was generated with `--schema-bearer-token-secret {{ .SchemaBearerTokenSecret.Name }}:{{ .SchemaBearerTokenSecret.Key }}`.
{{- else }}

To scaffold that env var during init, add `--schema-bearer-token-secret pulumi-schema-auth:accessToken`.

For example:

```bash
kratix init {{ .SubCommand }} {{ .Name }} --schema <schema-url> --schema-bearer-token-secret pulumi-schema-auth:accessToken --group {{ .Group }} --kind {{ .Kind }}
```
{{- end }}

If you are updating an existing Promise manually, add this env var to the `{{ .PulumiGeneratorName }}` container:

```yaml
- name: PULUMI_ACCESS_TOKEN
  valueFrom:
    secretKeyRef:
      name: pulumi-schema-auth
      key: accessToken
```
{{ end }}


## Updating Dependencies

TBD
