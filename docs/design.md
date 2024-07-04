# Design

## Overall syntax

```
kratix [VERB] [OBJECT] [NAME] [FLAGS]
```

where `VERB`, `TYPE`, `NAME`, and `FLAGS` are:

- `VERB`: Specifies the operation that you want to perform, for example `init`, `update` or `add`. `Update` indicates update and delete existing values are possible, whereas `add` suggest this operation is append-only.
- `OBJECT`: Specifies the object to run the operation on, for example `promise`, `api`, `dependencies` and `container`.
- `NAME`: Specifies the name of the object. Some command may have name omitted, such as `kratix update api|dependencies` because each promise can only have one `api` and `dependencies` section.
- `FLAGS`: Specifies flags for the command. Preferably, all flags should have a shorter name. For example `--property FIELDNAME` can also be `--property=FIELDNAME` or `-p FIELDNAME`.

## Commands

### help

```
kratix help
```

### init

```bash
kratix init promise PROMISENAME --group myorg.com --kind database [--version v1] [--plural postgreses] [--split]
```

### update api

```
kratix update api --property FIELDNAME:FIELDTYPE [-p FIELDNAME:FIELDTYPE] [--property FIELDNAME-] [--group myorg.com] [--kind database] [--plural postgreses]
```

### add container workflow

```
kratix add container resource/configure/PIPELINENAME --image syntasso/postgres-resource:v1.0.0 [--name CONTAINERNAME]
```

### update dependencies

```
kratix update dependencies PATH-TO-LOCAL-DIR
```

### init from helm

```
kratix init helm-promise PROMISENAME --group myorg.com --kind database [--version v1] [--plural postgreses] --values-file PATHTO-VALUES-FILE --url chartURL
```

### init from operator

```
kratix init operator-promise PROMISENAME --group myorg.com --kind database [--version v1] [--plural postgreses] --operator-manifests PATH-TO-OPERATOR-RELEASE-MANIFEST --api-from CRD-FULLNAME(needs to exist in operator release manifest)
```
