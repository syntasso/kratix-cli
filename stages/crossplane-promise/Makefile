STAGE_TAG ?= "ghcr.io/syntasso/kratix-cli/from-api-to-crossplane-claim"
STAGE_VERSION ?= "dev"

.PHONY: test
test: # Run tests
	go run github.com/onsi/ginkgo/v2/ginkgo -r
	
build:
	docker build \
		--tag ${STAGE_TAG}:${STAGE_VERSION} \
		--tag ${STAGE_TAG}:latest \
		--file Dockerfile \
		.

build-and-push:
	docker buildx build \
		--builder kratix-cli-image-builder \
		--push \
		--platform linux/arm64,linux/amd64\
		--tag ${STAGE_TAG}:${STAGE_VERSION} \
		--tag ${STAGE_TAG}:latest \
		--file Dockerfile \
		.

build-and-load: build
	kind load docker-image ${STAGE_TAG}:${STAGE_VERSION} --name platform