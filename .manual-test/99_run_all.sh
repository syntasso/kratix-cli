#!/usr/bin/env bash
set -euo pipefail

# Runs the current manual regression suite for component-to-crd Task 02 behavior.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

"$SCRIPT_DIR/00_build_binary.sh"
"$SCRIPT_DIR/01_success_valid_schema.sh"
"$SCRIPT_DIR/02_error_multiple_components_without_component.sh"
"$SCRIPT_DIR/03_success_explicit_component_multi_schema.sh"
"$SCRIPT_DIR/04_error_unknown_component.sh"
"$SCRIPT_DIR/05_error_unsupported_construct.sh"
"$SCRIPT_DIR/06_error_malformed_schema_preflight.sh"

echo "manual tests: PASS"
