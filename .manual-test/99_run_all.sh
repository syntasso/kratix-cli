#!/usr/bin/env bash
set -euo pipefail

# Runs the manual regression suite.
# Discovers and runs all executable test scripts named [0-9][0-9]_*.sh.
# Set RUN_INTERNET_TESTS=1 to include scripts marked with REQUIRES_INTERNET=1.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
THIS_SCRIPT="$(basename "${BASH_SOURCE[0]}")"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Build the manual-test binary before running any test scripts.
"$REPO_DIR/scripts/build_binary" >/dev/null

for test_script in "$SCRIPT_DIR"/[0-9][0-9]_*.sh; do
  test_name="$(basename "$test_script")"
  if [[ "$test_name" == "$THIS_SCRIPT" ]]; then
    continue
  fi
  if [[ ! -x "$test_script" ]]; then
    continue
  fi

  if grep -q '^# REQUIRES_INTERNET=1$' "$test_script"; then
    if [[ "${RUN_INTERNET_TESTS:-0}" != "1" ]]; then
      continue
    fi
  fi

  "$test_script"
done

echo "manual tests: PASS"
