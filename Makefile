OPERATOR_ASPECT_TAG ?= "ghcr.io/syntasso/kratix-cli/from-api-to-operator"
HELM_ASPECT_TAG ?= "ghcr.io/syntasso/kratix-cli/helm-resource-configure"
CROSSPLANE_ASPECT_TAG ?= "ghcr.io/syntasso/kratix-cli/from-api-to-crossplane-claim"
TERRAFORM_MODULE_TAG ?= "ghcr.io/syntasso/kratix-cli/terraform-generate"
KRATIX_CLI_VERSION ?= "v0.1.0"

all: test build

.PHONY: test
test: # Run tests
	./aspects/helm-promise/pipeline_test.bash
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

build-aspects: build-operator-promise-aspect build-helm-promise-aspect build-terraform-module-promise-aspect

build-and-push-aspects: # build and push all aspects
	if ! docker buildx ls | grep -q "kratix-cli-image-builder"; then \
		docker buildx create --name kratix-cli-image-builder; \
	fi;
	make build-and-push-operator-promise-aspect
	make build-and-push-helm-promise-aspect
	make build-and-push-terraform-module-promise-aspect

.PHONY: help
help: # Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

build-operator-promise-aspect:
	docker build \
		--tag ${OPERATOR_ASPECT_TAG}:${KRATIX_CLI_VERSION} \
		--tag ${OPERATOR_ASPECT_TAG}:latest \
		--file aspects/operator-promise/Dockerfile \
		.

build-and-push-operator-promise-aspect:
	docker buildx build \
		--builder kratix-cli-image-builder \
		--push \
		--platform linux/arm64,linux/amd64\
		--tag ${OPERATOR_ASPECT_TAG}:${KRATIX_CLI_VERSION} \
		--tag ${OPERATOR_ASPECT_TAG}:latest \
		--file aspects/operator-promise/Dockerfile \
		.

build-helm-promise-aspect:
	docker build \
		--tag ${HELM_ASPECT_TAG}:${KRATIX_CLI_VERSION} \
		--tag ${HELM_ASPECT_TAG}:latest \
		--file aspects/helm-promise/Dockerfile \
		.

build-and-push-helm-promise-aspect:
	docker buildx build \
		--builder kratix-cli-image-builder \
		--push \
		--platform linux/arm64,linux/amd64\
		--tag ${HELM_ASPECT_TAG}:${KRATIX_CLI_VERSION} \
		--tag ${HELM_ASPECT_TAG}:latest \
		--file aspects/helm-promise/Dockerfile \
		aspects/helm-promise

build-crossplane-promise-aspect:
	docker build \
		--tag ${CROSSPLANE_ASPECT_TAG}:${KRATIX_CLI_VERSION} \
		--tag ${CROSSPLANE_ASPECT_TAG}:latest \
		--file aspects/crossplane-promise/Dockerfile \
		.

build-and-push-crossplane-promise-aspect:
	docker buildx build \
		--builder kratix-cli-image-builder \
		--push \
		--platform linux/arm64,linux/amd64\
		--tag ${CROSSPLANE_ASPECT_TAG}:${KRATIX_CLI_VERSION} \
		--tag ${CROSSPLANE_ASPECT_TAG}:latest \
		--file aspects/crossplane-promise/Dockerfile \
		aspects/crossplane-promise

build-and-load-crossplane-promise-aspect: build-crossplane-promise-aspect
	kind load docker-image ${CROSSPLANE_ASPECT_TAG}:${KRATIX_CLI_VERSION} --name platform

build-terraform-module-promise-aspect:
	docker build \
		--tag ${TERRAFORM_MODULE_TAG}:${KRATIX_CLI_VERSION} \
		--tag ${TERRAFORM_MODULE_TAG}:latest \
		--file aspects/terraform-module-promise/Dockerfile \
		.

build-and-push-terraform-module-promise-aspect:
	docker buildx build \
		--builder kratix-cli-image-builder \
		--push \
		--platform linux/arm64,linux/amd64\
		--tag ${TERRAFORM_MODULE_TAG}:${KRATIX_CLI_VERSION} \
		--tag ${TERRAFORM_MODULE_TAG}:latest \
		--file aspects/terraform-module-promise/Dockerfile \
		.

build-and-load-terraform-module-promise-aspect: build-terraform-module-promise-aspect
	kind load docker-image ${TERRAFORM_MODULE_TAG}:${KRATIX_CLI_VERSION} --name platform


release: check-version-alignment
	goreleaser release --prepare --clean --config .goreleaser.yaml

