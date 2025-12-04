STAGE_TAG ?= ghcr.io/syntasso/kratix-cli/from-api-to-crossplane-claim
STAGE_VERSION ?= dev
DRY_RUN ?= 0
# There are dependencies on code elsewhere in this repo
BASE_PATH ?= ../..

PUSH_FLAG := $(if $(filter 0,$(DRY_RUN)),--push,)

.PHONY: test
test:
	go run github.com/onsi/ginkgo/v2/ginkgo -r
	
build:
	docker build \
		--tag ${STAGE_TAG}:${STAGE_VERSION} \
		--tag ${STAGE_TAG}:latest \
		--file Dockerfile \
		${BASE_PATH}

build-and-push:
	if ! docker buildx ls | grep -q "kratix-cli-image-builder"; then \
		docker buildx create --name kratix-cli-image-builder; \
	fi;
	docker buildx build \
		--builder kratix-cli-image-builder \
		$(PUSH_FLAG) \
		--platform linux/arm64,linux/amd64\
		--tag ${STAGE_TAG}:${STAGE_VERSION} \
		--tag ${STAGE_TAG}:latest \
		--file Dockerfile \
		${BASE_PATH}

build-and-load: build
	kind load docker-image ${STAGE_TAG}:${STAGE_VERSION} --name platform