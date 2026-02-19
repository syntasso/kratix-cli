#!/usr/bin/env bash
set -euo pipefail

# End-to-end manual test:
# 1) extract Pulumi schema via `pulumi package get-schema`,
# 2) run pulumi-component-to-crd on a chosen component token,
# 3) validate that CRD YAML is emitted.
#
# This script is parameterized so the same flow can be reused across multiple
# local and internet-backed package examples.

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/helpers.sh"

ensure_bin

usage() {
  cat <<'USAGE'
Usage:
  07_e2e_pulumi_schema_to_crd.sh [options]

Options:
  --component <token>              Pulumi component token.
                                   Default: nodejs-component-provider:index:MyComponent
  --schema-source <source>         Pulumi schema source for `pulumi package get-schema`.
                                   Default: pulumi/tests/integration/component_provider/nodejs/component-provider-host/provider
  --work-name <name>               Optional artifact directory suffix.
  --package-name <name>            Optional expected package name in extracted schema.
  --install-plugin <k:n:v>         Optional plugin install before schema extraction.
                                   Repeatable, format: kind:name:version (example: resource:random:v4.18.0).
  --expect-schema-contains <text>  Optional schema assertion. Repeatable.
  --expect-crd-contains <text>     Optional CRD assertion. Repeatable.
  --skip-install                   Skip `pulumi install` for local directory sources.
  --help                           Show this help.

Examples:
  ./e2e_pulumi_schema_to_crd.sh

  ./e2e_pulumi_schema_to_crd.sh \
    --component nodejs-component-provider:index:MyComponent \
    --schema-source pulumi/tests/integration/component_provider/nodejs/component-provider-host/provider \
    --install-plugin resource:random:v4.18.0 \
    --expect-crd-contains aComplexTypeInput:

  ./e2e_pulumi_schema_to_crd.sh \
    --component eks:index:Cluster \
    --schema-source eks@4.2.0
USAGE
}

require_command() {
  local name="$1"
  command -v "$name" >/dev/null 2>&1 || {
    echo "missing required command: $name" >&2
    exit 1
  }
}

sanitize_token_for_name() {
  local token="$1"
  local sanitized
  sanitized="$(printf '%s' "$token" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//; s/-+/-/g')"
  if [[ -z "$sanitized" ]]; then
    echo "component"
  else
    echo "$sanitized"
  fi
}

DEFAULT_COMPONENT_TOKEN="nodejs-component-provider:index:MyComponent"
DEFAULT_SCHEMA_SOURCE="pulumi/tests/integration/component_provider/nodejs/component-provider-host/provider"
COMPONENT_TOKEN="$DEFAULT_COMPONENT_TOKEN"
SCHEMA_SOURCE="$DEFAULT_SCHEMA_SOURCE"
WORK_NAME=""
PACKAGE_NAME=""
SKIP_INSTALL="0"

PLUGIN_SPECS=()
EXPECT_SCHEMA_CONTAINS=()
EXPECT_CRD_CONTAINS=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --component)
      COMPONENT_TOKEN="${2:-}"
      shift 2
      ;;
    --schema-source)
      SCHEMA_SOURCE="${2:-}"
      shift 2
      ;;
    --work-name)
      WORK_NAME="${2:-}"
      shift 2
      ;;
    --package-name)
      PACKAGE_NAME="${2:-}"
      shift 2
      ;;
    --install-plugin)
      PLUGIN_SPECS+=("${2:-}")
      shift 2
      ;;
    --expect-schema-contains)
      EXPECT_SCHEMA_CONTAINS+=("${2:-}")
      shift 2
      ;;
    --expect-crd-contains)
      EXPECT_CRD_CONTAINS+=("${2:-}")
      shift 2
      ;;
    --skip-install)
      SKIP_INSTALL="1"
      shift
      ;;
    --help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ ${#PLUGIN_SPECS[@]} -eq 0 && "$COMPONENT_TOKEN" == "$DEFAULT_COMPONENT_TOKEN" && "$SCHEMA_SOURCE" == "$DEFAULT_SCHEMA_SOURCE" ]]; then
  # Default local provider fixture depends on the random plugin.
  PLUGIN_SPECS=("resource:random:v4.18.0")
fi

if [[ -z "$WORK_NAME" ]]; then
  WORK_NAME="$(sanitize_token_for_name "$COMPONENT_TOKEN")"
fi

require_command pulumi

WORKSPACE_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
if [[ -d "$SCHEMA_SOURCE" ]]; then
  SOURCE_KIND="dir"
  SOURCE_DIR="$(cd "$SCHEMA_SOURCE" && pwd)"
elif [[ -d "$WORKSPACE_DIR/$SCHEMA_SOURCE" ]]; then
  SOURCE_KIND="dir"
  SOURCE_DIR="$(cd "$WORKSPACE_DIR/$SCHEMA_SOURCE" && pwd)"
else
  SOURCE_KIND="package"
  SOURCE_DIR="$WORKSPACE_DIR"
fi

WORK_DIR="$WORK_ROOT_DIR/e2e.$WORK_NAME"
SCHEMA_PATH="$WORK_DIR/$WORK_NAME.schema.json"
CRD_PATH="$WORK_DIR/$WORK_NAME.crd.yaml"
PULUMI_INSTALL_LOG="$WORK_DIR/pulumi.install.log"
PULUMI_PLUGIN_INSTALL_LOG="$WORK_DIR/pulumi.plugin-install.log"
PULUMI_SCHEMA_STDERR="$WORK_DIR/pulumi.get-schema.stderr.log"
CRD_STDERR="$WORK_DIR/pulumi-component-to-crd.stderr.log"

mkdir -p "$WORK_DIR"

# Step 1: install local package dependencies when schema source is a directory.
if [[ "$SOURCE_KIND" == "dir" && "$SKIP_INSTALL" != "1" ]]; then
  (
    cd "$SOURCE_DIR"
    pulumi install >"$PULUMI_INSTALL_LOG" 2>&1
  )
fi

# Step 2: install any explicitly requested plugins.
: >"$PULUMI_PLUGIN_INSTALL_LOG"
for plugin_spec in "${PLUGIN_SPECS[@]}"; do
  IFS=':' read -r plugin_kind plugin_name plugin_version <<<"$plugin_spec"
  if [[ -z "$plugin_kind" || -z "$plugin_name" || -z "$plugin_version" ]]; then
    echo "invalid --install-plugin value: $plugin_spec (expected kind:name:version)" >&2
    exit 1
  fi

  (
    cd "$SOURCE_DIR"
    pulumi plugin install "$plugin_kind" "$plugin_name" "$plugin_version" >>"$PULUMI_PLUGIN_INSTALL_LOG" 2>&1
  )
done

# Step 3: extract Pulumi schema JSON.
if [[ "$SOURCE_KIND" == "dir" ]]; then
  (
    cd "$SOURCE_DIR"
    pulumi package get-schema . >"$SCHEMA_PATH" 2>"$PULUMI_SCHEMA_STDERR"
  )
else
  (
    cd "$SOURCE_DIR"
    pulumi package get-schema "$SCHEMA_SOURCE" >"$SCHEMA_PATH" 2>"$PULUMI_SCHEMA_STDERR"
  )
fi

assert_file_contains "$SCHEMA_PATH" '"isComponent": true'
assert_file_contains "$SCHEMA_PATH" "$COMPONENT_TOKEN"
if [[ -n "$PACKAGE_NAME" ]]; then
  assert_file_contains "$SCHEMA_PATH" "\"name\": \"$PACKAGE_NAME\""
fi
for pattern in "${EXPECT_SCHEMA_CONTAINS[@]}"; do
  assert_file_contains "$SCHEMA_PATH" "$pattern"
done

# Step 4: convert schema to a CRD scaffold.
if run_capture "$CRD_PATH" "$CRD_STDERR" --in "$SCHEMA_PATH" --component "$COMPONENT_TOKEN"; then
  code=0
else
  code=$?
fi

assert_eq "$code" "0" "exit code"
assert_empty_file "$CRD_STDERR"
assert_file_contains "$CRD_PATH" 'kind: CustomResourceDefinition'
assert_file_contains "$CRD_PATH" 'openAPIV3Schema:'
assert_file_contains "$CRD_PATH" "name: \"$(sanitize_token_for_name "$COMPONENT_TOKEN")\""
for pattern in "${EXPECT_CRD_CONTAINS[@]}"; do
  assert_file_contains "$CRD_PATH" "$pattern"
done

echo "e2e manual test: PASS"
echo "schema source: $SCHEMA_SOURCE"
echo "component:     $COMPONENT_TOKEN"
echo "schema:        $SCHEMA_PATH"
echo "crd:           $CRD_PATH"
