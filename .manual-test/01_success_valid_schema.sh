#!/usr/bin/env bash
set -euo pipefail

# Success path: single component auto-select with translated spec schema.

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
ensure_bin

schema="$SCRIPT_DIR/schema.valid.json"
stdout="$SCRIPT_DIR/out.success.stdout.txt"
stderr="$SCRIPT_DIR/out.success.stderr.txt"

cat >"$schema" <<'JSON'
{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"name":{"type":"string"},"replicas":{"type":"integer","default":2}},"requiredInputs":["name"]}}}
JSON

if run_capture "$stdout" "$stderr" --in "$schema"; then code=0; else code=$?; fi
assert_eq "$code" "0" "exit code"
assert_empty_file "$stderr"
assert_file_contains "$stdout" 'apiVersion: apiextensions.k8s.io/v1'
assert_file_contains "$stdout" 'kind: CustomResourceDefinition'
assert_file_contains "$stdout" 'openAPIV3Schema:'
assert_file_contains "$stdout" 'required:'
assert_file_contains "$stdout" 'default: 2'
