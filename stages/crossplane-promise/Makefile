STAGE_TAG ?= ghcr.io/syntasso/kratix-cli/from-api-to-crossplane-claim
STAGE_VERSION ?= dev
DRY_RUN ?= 0
# There are dependencies on code elsewhere in this repo
BASE_PATH ?= ../..

PUSH_FLAG := $(if $(filter 0,$(DRY_RUN)),--push,)

.PHONY: help
help: # Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

.PHONY: test
test: # Run tests
	go run github.com/onsi/ginkgo/v2/ginkgo -r
	
build: # Build local container image
	docker build \
		--tag ${STAGE_TAG}:${STAGE_VERSION} \
		--tag ${STAGE_TAG}:latest \
		--file Dockerfile \
		${BASE_PATH}

build-and-push: # Build container image and push it to the container registry
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

build-and-load: build # Build container image and load it into kind
	kind load docker-image ${STAGE_TAG}:${STAGE_VERSION} --name platform