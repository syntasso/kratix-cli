#!/usr/bin/env bash
set -euo pipefail

# Scenario: valid --in argument with a single component schema.
# Expected stdout: scaffold YAML for pkg:index:Thing
# Expected stderr: (empty)
# Expected exit code: 0

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_PATH="$SCRIPT_DIR/component-to-crd"
SCHEMA_PATH="$SCRIPT_DIR/schema.valid.json"

if [[ ! -x "$BIN_PATH" ]]; then
  "$SCRIPT_DIR/00_build_binary.sh"
fi

cat > "$SCHEMA_PATH" <<'JSON'
{"resources":{"pkg:index:Thing":{"isComponent":true}}}
JSON

STDOUT_PATH="$SCRIPT_DIR/out.success.stdout.txt"
STDERR_PATH="$SCRIPT_DIR/out.success.stderr.txt"

set +e
"$BIN_PATH" --in "$SCHEMA_PATH" >"$STDOUT_PATH" 2>"$STDERR_PATH"
EXIT_CODE=$?
set -e

if [[ $EXIT_CODE -ne 0 ]]; then
  echo "unexpected exit code: $EXIT_CODE (expected 0)" >&2
  exit 1
fi

ACTUAL_STDOUT="$(cat "$STDOUT_PATH")"
if ! grep -q 'apiVersion: apiextensions.k8s.io/v1' "$STDOUT_PATH"; then
  echo "stdout missing apiVersion, got: $ACTUAL_STDOUT" >&2
  exit 1
fi

if ! grep -q 'kind: CustomResourceDefinition' "$STDOUT_PATH"; then
  echo "stdout missing CRD kind, got: $ACTUAL_STDOUT" >&2
  exit 1
fi

if ! grep -q 'placeholder scaffold for pkg:index:Thing' "$STDOUT_PATH"; then
  echo "stdout missing placeholder descriptor, got: $ACTUAL_STDOUT" >&2
  exit 1
fi

if [[ -s "$STDERR_PATH" ]]; then
  echo "expected empty stderr, got:" >&2
  cat "$STDERR_PATH" >&2
  exit 1
fi
