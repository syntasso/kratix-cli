# Kratix CLI

The best tool you'll ever find to build your Promises!

## Installation
To build the CLI, run:

```bash
make build
```
The binary will be available at `./bin/kratix`.

## Usage

### Initializing promise

To bootstrap the Promise, you can use `kratix init promise` command:
```
kratix init promise PROMISE-NAME --group API-GROUP --kind API-KIND [--version] [--plural]
```

### Updating API properties

To update the Promise API, you can use the `kratix update api` command:

```
kratix update api --property PROPERTY-NAME:string -p PROPERTY-NAME:number [-p PROPERTY-NAME-] [--kind]
```

### Updating Workflows

To add workflow containers, you can use the `kratix add container` command:

```
kratix add container WORKFLOW/ACTION/PIPELINENAME --image CONTAINER-IMAGE [--name]
```

To see helpful messages about using the cli, you can run:
```
kratix help
kratix help init
kratix help update api
kratix add container --help
```

## Testing

To run the tests, run:

```bash
make test
```
