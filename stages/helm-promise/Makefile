HELM_STAGE_TAG ?= "ghcr.io/syntasso/kratix-cli/helm-resource-configure"
HELM_STAGE_VERSION ?= "dev"

.PHONY: test
test: # Run tests
	./pipeline_test.bash

build-helm-promise-stage:
	docker build \
		--tag ${HELM_STAGE_TAG}:${HELM_STAGE_VERSION} \
		--tag ${HELM_STAGE_TAG}:latest \
		--file stages/helm-promise/Dockerfile \
		.

build-and-push-helm-promise-stage:
	docker buildx build \
		--builder kratix-cli-image-builder \
		--push \
		--platform linux/arm64,linux/amd64\
		--tag ${HELM_STAGE_TAG}:${HELM_STAGE_VERSION} \
		--tag ${HELM_STAGE_TAG}:latest \
		--file stages/helm-promise/Dockerfile \
		stages/helm-promise

build-crossplane-promise-stage:
	docker build \
		--tag ${CROSSPLANE_STAGE_TAG}:${CROSSPLANE_STAGE_VERSION} \
		--tag ${CROSSPLANE_STAGE_TAG}:latest \
		--file stages/crossplane-promise/Dockerfile \
		.

