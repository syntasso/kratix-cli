import kratix_sdk as ks

sdk = ks.KratixSDK()
if workflow_action() == "promise":
	print(f'Hello from {sdk.promise_name()}')
else:
	resource = sdk.read_resource_input()
	print(f'Hello from {resource.get_name()} {resource.get_namespace()}')