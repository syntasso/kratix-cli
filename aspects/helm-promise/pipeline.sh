#!/usr/bin/env sh

set -eux

KRATIX_INPUT=${KRATIX_INPUT:-/kratix/input}
KRATIX_OUTPUT=${KRATIX_OUTPUT:-/kratix/output}
HELM_BINARY=${HELM_BINARY:-helm}
name=$(yq '.metadata.name' $KRATIX_INPUT/object.yaml)


yq '.spec' --output-format yaml $KRATIX_INPUT/object.yaml > values.yaml

if [ -n "$CHART_URL" ]; then
  if [ -n "${CHART_NAME:-}" ]; then
    arguments="$CHART_NAME --repo $CHART_URL"
  else
    arguments="$CHART_URL"
  fi
else
    echo "URL is not set. Please set the URL."
    exit 1
fi


if [ -n "${CHART_VERSION:-}" ]; then
    arguments="$arguments --version $CHART_VERSION"
fi

$HELM_BINARY template $name $arguments --values values.yaml > $KRATIX_OUTPUT/object.yaml
