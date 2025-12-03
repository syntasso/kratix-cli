IMG_TAG ?= "ghcr.io/syntasso/kratix-cli/helm-resource-configure"
VERSION ?= "dev"

.PHONY: test
test: # Run tests
	./pipeline_test.bash

build:
	docker build \
		--tag ${IMG_TAG}:${VERSION} \
		--tag ${IMG_TAG}:latest \
		--file Dockerfile \
		.

build-and-push:
	docker buildx build \
		--builder kratix-cli-image-builder \
		--push \
		--platform linux/arm64,linux/amd64\
		--tag ${IMG_TAG}:${VERSION} \
		--tag ${IMG_TAG}:latest \
		--file Dockerfile \
		.

build-and-load: build
	kind load docker-image ${IMG_TAG}:${VERSION} --name platform