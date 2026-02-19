#!/usr/bin/env bash
set -euo pipefail

# Scenario: schema has multiple components and --component is omitted.
# Expected stderr contains sorted available tokens.
# Expected exit code: 2

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_PATH="$SCRIPT_DIR/component-to-crd"
SCHEMA_PATH="$SCRIPT_DIR/schema.multi-components.json"

if [[ ! -x "$BIN_PATH" ]]; then
  "$SCRIPT_DIR/00_build_binary.sh"
fi

cat > "$SCHEMA_PATH" <<'JSON'
{"resources":{"pkg:index:Zulu":{"isComponent":true},"pkg:index:Alpha":{"isComponent":true},"pkg:index:Skip":{"isComponent":false}}}
JSON

set +e
OUTPUT="$($BIN_PATH --in "$SCHEMA_PATH" 2>&1)"
EXIT_CODE=$?
set -e

if [[ $EXIT_CODE -ne 2 ]]; then
  echo "unexpected exit code: $EXIT_CODE (expected 2)" >&2
  exit 1
fi

EXPECTED='error: multiple components found; provide --component from: pkg:index:Alpha, pkg:index:Zulu'
if [[ "$OUTPUT" != "$EXPECTED" ]]; then
  echo "unexpected output: $OUTPUT" >&2
  echo "expected output: $EXPECTED" >&2
  exit 1
fi

echo "$OUTPUT"
