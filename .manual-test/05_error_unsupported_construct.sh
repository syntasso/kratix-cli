#!/usr/bin/env bash
set -euo pipefail

# Error path: unsupported schema construct should return exit 3.

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
ensure_bin

schema="$SCRIPT_DIR/schema.unsupported.json"
cat >"$schema" <<'JSON'
{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"value":{"oneOf":[{"type":"string"},{"type":"number"}]}}}}}
JSON

if out="$(run_combined --in "$schema")"; then code=0; else code=$?; fi
assert_eq "$code" "3" "exit code"
assert_eq "$out" 'error: component "pkg:index:Thing" path "spec.value" unsupported construct: keyword "oneOf"' "stderr"
echo "$out"
