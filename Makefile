OPERATOR_STAGE_TAG ?= "ghcr.io/syntasso/kratix-cli/from-api-to-operator"
HELM_STAGE_TAG ?= "ghcr.io/syntasso/kratix-cli/helm-resource-configure"
CROSSPLANE_STAGE_TAG ?= "ghcr.io/syntasso/kratix-cli/from-api-to-crossplane-claim"
TERRAFORM_MODULE_TAG ?= "ghcr.io/syntasso/kratix-cli/terraform-generate"
KRATIX_CLI_VERSION ?= "v0.2.0"
TERRAFORM_STAGE_VERSION ?= "v0.3.0"

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

build-stages: build-operator-promise-stage build-helm-promise-stage build-terraform-module-promise-stage

build-and-push-stages: # build and push all stages
	if ! docker buildx ls | grep -q "kratix-cli-image-builder"; then \
		docker buildx create --name kratix-cli-image-builder; \
	fi;
	make build-and-push-operator-promise-stage
	make build-and-push-helm-promise-stage
	make build-and-push-terraform-module-promise-stage

.PHONY: help
help: # Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

build-operator-promise-stage:
	docker build \
		--tag ${OPERATOR_STAGE_TAG}:${KRATIX_CLI_VERSION} \
		--tag ${OPERATOR_STAGE_TAG}:latest \
		--file stages/operator-promise/Dockerfile \
		.

build-and-push-operator-promise-stage:
	docker buildx build \
		--builder kratix-cli-image-builder \
		--push \
		--platform linux/arm64,linux/amd64\
		--tag ${OPERATOR_STAGE_TAG}:${KRATIX_CLI_VERSION} \
		--tag ${OPERATOR_STAGE_TAG}:latest \
		--file stages/operator-promise/Dockerfile \
		.

build-helm-promise-stage:
	docker build \
		--tag ${HELM_STAGE_TAG}:${KRATIX_CLI_VERSION} \
		--tag ${HELM_STAGE_TAG}:latest \
		--file stages/helm-promise/Dockerfile \
		.

build-and-push-helm-promise-stage:
	docker buildx build \
		--builder kratix-cli-image-builder \
		--push \
		--platform linux/arm64,linux/amd64\
		--tag ${HELM_STAGE_TAG}:${KRATIX_CLI_VERSION} \
		--tag ${HELM_STAGE_TAG}:latest \
		--file stages/helm-promise/Dockerfile \
		stages/helm-promise

build-crossplane-promise-stage:
	docker build \
		--tag ${CROSSPLANE_STAGE_TAG}:${KRATIX_CLI_VERSION} \
		--tag ${CROSSPLANE_STAGE_TAG}:latest \
		--file stages/crossplane-promise/Dockerfile \
		.

build-and-push-crossplane-promise-stage:
	docker buildx build \
		--builder kratix-cli-image-builder \
		--push \
		--platform linux/arm64,linux/amd64\
		--tag ${CROSSPLANE_STAGE_TAG}:${KRATIX_CLI_VERSION} \
		--tag ${CROSSPLANE_STAGE_TAG}:latest \
		--file stages/crossplane-promise/Dockerfile \
		stages/crossplane-promise

build-and-load-crossplane-promise-stage: build-crossplane-promise-stage
	kind load docker-image ${CROSSPLANE_STAGE_TAG}:${KRATIX_CLI_VERSION} --name platform

build-terraform-module-promise-stage:
	docker build \
		--tag ${TERRAFORM_MODULE_TAG}:${TERRAFORM_STAGE_VERSION} \
		--tag ${TERRAFORM_MODULE_TAG}:latest \
		--file stages/terraform-module-promise/Dockerfile \
		.

build-and-load-operator-promise-stage: build-operator-promise-stage
	kind load docker-image ${OPERATOR_STAGE_TAG}:${KRATIX_CLI_VERSION} --name platform

build-and-push-terraform-module-promise-stage:
	docker buildx build \
		--builder kratix-cli-image-builder \
		--push \
		--platform linux/arm64,linux/amd64\
		--tag ${TERRAFORM_MODULE_TAG}:${TERRAFORM_STAGE_VERSION} \
		--tag ${TERRAFORM_MODULE_TAG}:latest \
		--file stages/terraform-module-promise/Dockerfile \
		.

build-and-load-terraform-module-promise-stage: build-terraform-module-promise-stage
	kind load docker-image ${TERRAFORM_MODULE_TAG}:${TERRAFORM_STAGE_VERSION} --name platform


release: check-version-alignment
	goreleaser release --prepare --clean --config .goreleaser.yaml
