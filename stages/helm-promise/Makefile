IMG_TAG ?= "ghcr.io/syntasso/kratix-cli/helm-resource-configure"
VERSION ?= "dev"
DRY_RUN ?= 0

PUSH_FLAG := $(if $(filter 0,$(DRY_RUN)),--push,)

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
	if ! docker buildx ls | grep -q "kratix-cli-image-builder"; then \
		docker buildx create --name kratix-cli-image-builder; \
	fi;
	docker buildx build \
		--builder kratix-cli-image-builder \
		$(PUSH_FLAG) \
		--platform linux/arm64,linux/amd64\
		--tag ${IMG_TAG}:${VERSION} \
		--tag ${IMG_TAG}:latest \
		--file Dockerfile \
		.

build-and-load: build
	kind load docker-image ${IMG_TAG}:${VERSION} --name platform
