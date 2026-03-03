|
  # Promise Template

  This Promise was generated with:

  ```
  kratix init pulumi-component-promise mypromise --schema './schema.valid.json' --group syntasso.io --kind Database
  ```

  ## Updating API properties

  To update the Promise API, you can use the `kratix update api` command:

  ```
  kratix update api --property name:string --property region- --kind Database
  ```

  ## Updating Workflows

  To add workflow containers, you can use the `kratix add container` command:

  ```
  kratix add container resource/configure/pipeline0 --image syntasso/postgres-resource:v1.0.0
  ```


  ### Pulumi PKO output

  The generated Pulumi workflow runs two containers from the same stage codebase:
  - `/pulumi-program-generator` entrypoint emits a PKO `Program`.
  - `/pulumi-stack-generator` entrypoint emits a PKO `Stack`.

  The `Program` container writes deterministic values from existing inputs:
  - `program.resources.<component>.type` from `--component` selection.
  - `program.resources.<component>.properties` from request `spec`.
  - deterministic metadata, namespace and naming from the stage contract.

  The `Program` container also reads `PULUMI_SCHEMA_SOURCE` and auto-generates `program.configuration` entries only when values are explicitly trusted in schema config variables (`type`, `default`, `secret`).

  The `Stack` container writes deterministic metadata passthrough, `spec.programRef.name`, and `spec.stack` from known request inputs and component identity.


  This PKO object generation introduces no additional required environment variables for either the Program or Stack.
  If you need additional Pulumi runtime intent, write and add a custom stage container to the Workflow that updates the generated Program or Stack before it is written to stage output.


  ## Updating Dependencies

  TBD
