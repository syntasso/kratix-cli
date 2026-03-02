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

The generated Pulumi stage writes a PKO `Program` with deterministic values from existing inputs:
- `program.resources.<component>.type` from `--component` selection.
- `program.resources.<component>.properties` from request `spec`.
- deterministic metadata, namespace and naming from the stage contract.

The stage also reads `PULUMI_SCHEMA_SOURCE` and auto-generates `program.configuration` entries only when values are explicitly trusted in schema config variables (`type`, `default`, `secret`).

This Program generation introduces no additional required environment variables for PKO Program required fields.

If you need optional Pulumi runtime intent (for example resource options, variables, outputs, or environment-specific wiring), write a custom stage container that updates the generated Program before it is written to stage output.
{{ end }}


## Updating Dependencies

TBD
