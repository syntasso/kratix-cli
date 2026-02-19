#!/usr/bin/env bash
set -euo pipefail

# Error path: unknown component token.

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
ensure_bin

schema="$SCRIPT_DIR/schema.unknown-component.json"
cat >"$schema" <<'JSON'
{"resources":{"pkg:index:Zulu":{"isComponent":true},"pkg:index:Alpha":{"isComponent":true},"pkg:index:Skip":{"isComponent":false}}}
JSON

if out="$(run_combined --in "$schema" --component pkg:index:Missing)"; then code=0; else code=$?; fi
assert_eq "$code" "2" "exit code"
assert_eq "$out" 'error: component "pkg:index:Missing" not found; available components: pkg:index:Alpha, pkg:index:Zulu' "stderr"
echo "$out"
