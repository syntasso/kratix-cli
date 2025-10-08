# Kratix CLI

The best tool you'll ever find to build your Promises!

## Installation
To install the CLI, run:

```bash
go install github.com/syntasso/kratix-cli/cmd/kratix@latest
```

## Build
To build a dev version of the CLI, run:

```bash
make build
```

The binary will be available at `./bin/kratix`.

## Usage

### Initializing promise

To bootstrap the Promise, you can use `kratix init promise` command:
```
kratix init promise PROMISE-NAME --group API-GROUP --kind API-KIND [--version] [--plural] [--split]
```

### Updating API properties

To update the Promise API, you can use the `kratix update api` command:

```
kratix update api --property PROPERTY-NAME:string -p PROPERTY-NAME:number [-p PROPERTY-NAME-] [--kind]
```

### Updating Workflows

To add workflow containers, you can use the `kratix add container` command:

```
kratix add container WORKFLOW/ACTION/PIPELINENAME --image CONTAINER-IMAGE [--name] [--language]
```

### Updating Dependencies

To add Promise dependencies, you can run the `kratix update dependencies dependencies` command:
```
kratix update dependencies DEPENDENCIES-DIRECTORY/
```

### Updating Destination selectors

To update Destination selectors of the Promise, you can use the `kratix update destination-selector` command:
```
kratix update destination-selector env=dev
```

### Building Promise

If you initialized the Promise by providing `--split` flag in `kratix init promise` command, run
the `kratix build promise` command to combine the Promise api, workflow, and dependencies:
```
kratix build promise PROMISE-NAME
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

# Releasing

To release merge the auto-created Release PR
([example](https://github.com/syntasso/kratix-cli/pull/48)). This PR is auto
created by the [Release Please](https://github.com/googleapis/release-please)
Github Action we have in our `.github/workflows/release.yml` file. When this PR
is merged the following happens:

- A tag and Github release is created. The release notes is equal to the
   contents of the PRs description (**NOT the contents of the file committed**).
- Goreleaser gets triggered in Github actions, creating and uploading the binaries to the
   existing release.

