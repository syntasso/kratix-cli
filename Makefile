OPERATOR_ASPECT_TAG ?= "ghcr.io/syntasso/kratix-cli/from-api-to-operator"
HELM_ASPECT_TAG ?= "ghcr.io/syntasso/kratix-cli/helm-reosource-configure"
KRATIX_CLI_VERSION ?= "v0.1.0"

all: test build

.PHONY: test
test: # Run tests
	./aspects/helm-promise/pipeline_test.bash
	go run github.com/onsi/ginkgo/v2/ginkgo -r

build: # Build the binary
	go build -o bin/kratix main.go

build-aspects: build-operator-promise-aspect  # build all aspects

build-and-push-aspects: # build and push all aspects
	if ! docker buildx ls | grep -q "kratix-cli-image-builder"; then \
		docker buildx create --name kratix-cli-image-builder; \
	fi;
	make build-and-push-operator-promise-aspect
	make build-and-push-helm-promis-aspect

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
		aspects/helm-promise

build-and-push-helm-promise-aspect: build-helm-promise-aspect
	docker push ${HELM_ASPECT_TAG}:${KRATIX_CLI_VERSION}
	docker push ${HELM_ASPECT_TAG}:latest
