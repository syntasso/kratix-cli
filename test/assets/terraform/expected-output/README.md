# Promise Template

This Promise was generated with:

```
kratix init tf-module-promise googlecloudrun --module-source https://github.com/GoogleCloudPlatform/terraform-google-cloud-run --module-version v0.16.4 --group gcp.com --kind GoogleCloudRun
```

## Updating API properties

To update the Promise API, you can use the `kratix update api` command:

```
kratix update api --property name:string --property region- --kind GoogleCloudRun
```

## Updating Workflows

To add workflow containers, you can use the `kratix add container` command:

```
kratix add container resource/configure/pipeline0 --image syntasso/postgres-resource:v1.0.0
```

## Updating Dependencies

TBD
