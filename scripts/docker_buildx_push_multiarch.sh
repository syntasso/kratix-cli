#!/usr/bin/env bash
set -euo pipefail

# Build and push a multi-architecture Docker image for pulumi-component-to-crd.
# Required env:
# - IMAGE_TAG, e.g. ghcr.io/my-org/pulumi-component-to-crd:v0.1.0
# Optional env:
# - PLATFORMS (default: linux/amd64,linux/arm64)
# - BUILDER (default: current buildx builder)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required" >&2
  exit 1
fi

if ! docker buildx version >/dev/null 2>&1; then
  echo "docker buildx is required" >&2
  exit 1
fi

if [[ -z "${IMAGE_TAG:-}" ]]; then
  echo "missing required env IMAGE_TAG (example: ghcr.io/my-org/pulumi-component-to-crd:v0.1.0)" >&2
  exit 1
fi

platforms="${PLATFORMS:-linux/amd64,linux/arm64}"
builder="${BUILDER:-}"

cmd=(docker buildx build --platform "$platforms" -t "$IMAGE_TAG" --push)
if [[ -n "$builder" ]]; then
  cmd+=(--builder "$builder")
fi
cmd+=("$REPO_DIR")

echo "building and pushing image: $IMAGE_TAG ($platforms)"
"${cmd[@]}"
echo "pushed image: $IMAGE_TAG"
