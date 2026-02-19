#!/usr/bin/env bash
set -euo pipefail

# Build a local-architecture Docker image for pulumi-component-to-crd.
# Optional env:
# - IMAGE_TAG (default: pulumi-component-to-crd:local)
# - PLATFORM (default: linux/<host-arch>)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required" >&2
  exit 1
fi

image_tag="${IMAGE_TAG:-pulumi-component-to-crd:local}"
platform="${PLATFORM:-}"

if [[ -z "$platform" ]]; then
  case "$(uname -m)" in
    arm64|aarch64) platform="linux/arm64" ;;
    x86_64|amd64) platform="linux/amd64" ;;
    *)
      echo "unsupported host architecture: $(uname -m); set PLATFORM explicitly" >&2
      exit 1
      ;;
  esac
fi

echo "building image: $image_tag ($platform)"
docker build --platform "$platform" -t "$image_tag" "$REPO_DIR"
echo "built image: $image_tag"
