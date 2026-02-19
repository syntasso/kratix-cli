#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN_DIR="$(cd "$REPO_DIR/bin" && pwd)"
BIN_PATH="$REPO_DIR/bin/component-to-crd"

ensure_bin() {
  [[ -x "$BIN_PATH" ]] || "$SCRIPT_DIR/build_binary" >/dev/null
}

run_capture() {
  local stdout_path="$1"
  local stderr_path="$2"
  shift 2

  set +e
  "$BIN_PATH" "$@" >"$stdout_path" 2>"$stderr_path"
  local code=$?
  set -e
  return "$code"
}

run_combined() {
  set +e
  local out
  out="$($BIN_PATH "$@" 2>&1)"
  local code=$?
  set -e
  printf '%s' "$out"
  return "$code"
}

assert_eq() {
  local got="$1"
  local want="$2"
  local label="$3"
  [[ "$got" == "$want" ]] || { echo "$label mismatch: got [$got], want [$want]" >&2; exit 1; }
}

assert_file_contains() {
  local path="$1"
  local pattern="$2"
  grep -q -- "$pattern" "$path" || { echo "$path missing: $pattern" >&2; cat "$path" >&2; exit 1; }
}

assert_empty_file() {
  local path="$1"
  [[ ! -s "$path" ]] || { echo "$path expected empty" >&2; cat "$path" >&2; exit 1; }
}
