#!/usr/bin/env bash
set -euo pipefail

# Success path: explicit component selection in multi-component schema.

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
ensure_bin

schema="$SCRIPT_DIR/schema.with-component.json"
stdout="$SCRIPT_DIR/out.component-selected.stdout.txt"
stderr="$SCRIPT_DIR/out.component-selected.stderr.txt"

cat >"$schema" <<'JSON'
{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"name":{"type":"string"}}},"pkg:index:Other":{"isComponent":true,"inputProperties":{"other":{"type":"string"}}},"pkg:index:Skip":{"isComponent":false}}}
JSON

if run_capture "$stdout" "$stderr" --in "$schema" --component pkg:index:Thing; then code=0; else code=$?; fi
assert_eq "$code" "0" "exit code"
assert_empty_file "$stderr"
assert_file_contains "$stdout" 'openAPIV3Schema:'
assert_file_contains "$stdout" 'name:'
