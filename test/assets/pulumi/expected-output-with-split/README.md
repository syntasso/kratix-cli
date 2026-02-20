|
  # Promise Template

  This Promise was generated with:

  ```
  kratix init pulumi-component-promise mypromise --schema ./schema.valid.json --split --group syntasso.io --kind Database
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

  ## Updating Dependencies

  TBD
