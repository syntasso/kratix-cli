#!/usr/bin/env bash
set -euo pipefail

# Task 07 Docker packaging smoke test.
# Validates:
# 1) Docker image builds from component-to-crd/.
# 2) Containerized run against mounted schema succeeds.
# 3) Containerized stdout exactly matches local binary stdout.

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
"$SCRIPT_DIR/00_build_binary.sh" >/dev/null

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required for this smoke test" >&2
  exit 1
fi
if ! docker info >/dev/null 2>&1; then
  echo "docker daemon is not available" >&2
  exit 1
fi

work_dir="$SCRIPT_DIR/work.docker-smoke"
mkdir -p "$work_dir"
repo_dir="$(cd "$SCRIPT_DIR/.." && pwd)"

local_out="$work_dir/local.stdout.yaml"
docker_out="$work_dir/docker.stdout.yaml"
schema_path="$SCRIPT_DIR/schema.valid.json"

image_tag="component-to-crd:manual-smoke"

"$BIN_PATH" --in "$schema_path" >"$local_out"

IMAGE_TAG="$image_tag" "$repo_dir/scripts/docker_build_local.sh" >/dev/null

docker run --rm \
  --mount "type=bind,src=$SCRIPT_DIR,dst=/work,readonly" \
  "$image_tag" \
  --in /work/schema.valid.json >"$docker_out"

diff -u "$local_out" "$docker_out"

echo "docker smoke: PASS"
