#!/usr/bin/env bash
set -euo pipefail

# Error path: malformed schema should fail preflight with exit 2 before selection/translation.

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
ensure_bin

schema="$SCRIPT_DIR/schema.malformed.json"
stdout_file="$SCRIPT_DIR/out.malformed.stdout.txt"
stderr_file="$SCRIPT_DIR/out.malformed.stderr.txt"
stdout_file_unsupported="$SCRIPT_DIR/out.malformed-over-unsupported.stdout.txt"
stderr_file_unsupported="$SCRIPT_DIR/out.malformed-over-unsupported.stderr.txt"
cleanup() {
  local status=$?
  if [[ "$status" -eq 0 ]]; then
    rm -f "$schema" "$stdout_file" "$stderr_file" "$stdout_file_unsupported" "$stderr_file_unsupported"
  fi
}
trap cleanup EXIT

cat >"$schema" <<'JSON'
{"resources":{"pkg:index:Zulu":{"isComponent":true,"inputProperties":{"bad":{"$ref":"#/types/pkg:index:Missing"}}},"pkg:index:Alpha":{"isComponent":true,"inputProperties":{"value":{"oneOf":[{"type":"string"},{"type":"number"}]}}}}}
JSON

if run_capture "$stdout_file" "$stderr_file" --in "$schema" --component "pkg:index:Missing"; then code=0; else code=$?; fi
assert_eq "$code" "2" "exit code"
assert_empty_file "$stdout_file"
assert_file_contains "$stderr_file" 'error: schema preflight path'
assert_file_contains "$stderr_file" 'resources.pkg:index:Zulu.inputProperties.bad'
assert_file_contains "$stderr_file" 'unresolved local type ref "#/types/pkg:index:Missing"'
if grep -q -- 'component "pkg:index:Missing" not found' "$stderr_file"; then
  echo "unexpected selection error; preflight should win" >&2
  cat "$stderr_file" >&2
  exit 1
fi
if grep -q -- 'unsupported construct' "$stderr_file"; then
  echo "unexpected unsupported classification; preflight should win" >&2
  cat "$stderr_file" >&2
  exit 1
fi
if [[ "$(awk 'END{print NR}' "$stderr_file")" != "1" ]]; then
  echo "stderr should be single-line parseable output" >&2
  cat "$stderr_file" >&2
  exit 1
fi

# malformed preflight should also win over unsupported classification on a selected component
if run_capture "$stdout_file_unsupported" "$stderr_file_unsupported" --in "$schema" --component "pkg:index:Alpha"; then code=0; else code=$?; fi
assert_eq "$code" "2" "exit code (preflight over unsupported)"
assert_empty_file "$stdout_file_unsupported"
assert_file_contains "$stderr_file_unsupported" 'error: schema preflight path'
assert_file_contains "$stderr_file_unsupported" 'unresolved local type ref "#/types/pkg:index:Missing"'
if grep -q -- 'unsupported construct' "$stderr_file_unsupported"; then
  echo "unexpected unsupported classification; preflight should win" >&2
  cat "$stderr_file_unsupported" >&2
  exit 1
fi
if [[ "$(awk 'END{print NR}' "$stderr_file_unsupported")" != "1" ]]; then
  echo "stderr should be single-line parseable output" >&2
  cat "$stderr_file_unsupported" >&2
  exit 1
fi

cat "$stderr_file"
