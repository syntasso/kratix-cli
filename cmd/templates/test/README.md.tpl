# Test directory for container image `{{ .RawImageName }}`

This directory contains testcases for the Kratix container image `{{ .RawImageName }}`.

## Directory structure

Initially, the directory structure is as follows:

```
{{ .Directory }}/
	test_{{ .FormattedImageName }}/
```

As testcases are added, the directory structure is updated as follows:

```
{{ .Directory }}/
	test_{{ .FormattedImageName }}/
	  resource/
	    configure/
			  test_one/
			    before/
			      input/
			      output/
			      metadata/
			    after/
			      input/
			      output/
			      metadata/
			  ...
    promise/
      configure/
        test_two/
          before/
            input/
            output/
            metadata/
          after/
            input/
            output/
            metadata/
        ...
  ...
```

The `resource/` directory contains testcases for resource workflows, while the `promise/`
directory contains testcases for the promise workflows.

Each testcase directory contains the following subdirectories:
- `before/` - contains the **prior** state input, output and metadata directories **before** the container is run
- `after/` - contains the **expected** state of the input, output and metadata directories **after** the container is run

## Adding testcases

To add a new testcase, use the following command:

```
kratix test container add --image {{ .RawImageName }} <testcase-name>
```

To add a new testcase with an input object, use the following command:

```
kratix test container add --image {{ .RawImageName }} <testcase-name> --input-object <path-to-input-object>
```

## Running testcases

To run the testcases, use the following command:

```
kratix test container run --image {{ .RawImageName }}
```

To run specific testcases, use the following command:

```
kratix test container run --image {{ .RawImageName }} --testcases <testcase1>,<testcase2>,<testcase3>
```
