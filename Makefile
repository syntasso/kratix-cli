all: test build

.PHONY: test
test: # Run tests
	./stages/helm-promise/pipeline_test.bash
	go run github.com/onsi/ginkgo/v2/ginkgo -r

.PHONY: check-version-alignment
check-version-alignment:
	@manifest_version=$$(jq -r '."."' .release-please-manifest.json); \
	go_version=$$(awk -F'"' '/var version =/ {print $$2}' cmd/kratix/main.go); \
	if [ "$$manifest_version" != "$$go_version" ]; then \
		echo "❌ Version mismatch:"; \
		echo " - .release-please-manifest.json => $$manifest_version"; \
		echo " - cmd/kratix/main.go            => $$go_version"; \
		echo ""; \
		echo "💡 Please update main.go to match the release version."; \
		exit 1; \
	else \
		echo "✅ Versions are aligned: $$manifest_version"; \
	fi

build: # Build the binary
	CGO_ENABLED=0 go build -o bin/kratix ./cmd/kratix/main.go


.PHONY: help
help: # Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

# Build commands for stages
.PHONY: help
build-and-load-stages: build-and-load-crossplane-promise-stage build-and-load-helm-promise-stage build-and-load-operator-promise-stage build-and-load-terraform-module-promise-stage # Build container images for all stages and load them into kind

build-crossplane-promise-stage:
	$(MAKE) -C stages/crossplane-promise build

build-and-load-crossplane-promise-stage:
	$(MAKE) -C stages/crossplane-promise build-and-load

build-and-push-crossplane-promise-stage:
	$(MAKE) -C stages/crossplane-promise build-and-push

build-helm-promise-stage:
	$(MAKE) -C stages/helm-promise build

build-and-load-helm-promise-stage:
	$(MAKE) -C stages/helm-promise build-and-load

build-and-push-helm-promise-stage:
	$(MAKE) -C stages/helm-promise build-and-push

build-operator-promise-stage:
	$(MAKE) -C stages/operator-promise build

build-and-load-operator-promise-stage:
	$(MAKE) -C stages/operator-promise build-and-load

build-and-push-operator-promise-stage:
	$(MAKE) -C stages/operator-promise build-and-push

build-pulumi-promise-stage:
	$(MAKE) -C stages/pulumi-promise build

build-and-load-pulumi-promise-stage:
	$(MAKE) -C stages/pulumi-promise build-and-load

build-and-push-pulumi-promise-stage:
	$(MAKE) -C stages/pulumi-promise build-and-push

build-terraform-module-promise-stage:
	$(MAKE) -C stages/terraform-module-promise build

build-and-load-terraform-module-promise-stage:
	$(MAKE) -C stages/terraform-module-promise build-and-load

build-and-push-terraform-module-promise-stage:
	$(MAKE) -C stages/terraform-module-promise build-and-push

release: check-version-alignment
	goreleaser release --prepare --clean --config .goreleaser.yaml
