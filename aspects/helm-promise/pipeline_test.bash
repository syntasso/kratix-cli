#!/usr/bin/env bash

set -euo pipefail

ROOT=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
export HELM_BINARY=echo

function testOCI {
  echo "  testing OCI helm chart"
  export KRATIX_INPUT=/tmp/testOCI/kratix-input
  export KRATIX_OUTPUT=/tmp/testOCI/kratix-output
  mkdir -p $KRATIX_INPUT
  mkdir -p $KRATIX_OUTPUT

  cat <<EOF > "${KRATIX_INPUT}/object.yaml"
metadata:
  name: foo
spec:
  foo: bar
EOF

  CHART_URL=oci://registry-1.docker.io/bitnamicharts/redis CHART_VERSION=19.6.1 $ROOT/pipeline.sh 2>&1 | grep "template foo oci://registry-1.docker.io/bitnamicharts/redis --version 19.6.1 --values values.yaml"
  echo "  testing OCI helm chart passed"
  rm -rf $KRATIX_INPUT
  rm -rf $KRATIX_OUTPUT
}

function testRepo {
  echo "  testing helm chart from a repo with a name"
  export KRATIX_INPUT=/tmp/testRepo/kratix-input
  export KRATIX_OUTPUT=/tmp/testRepo/kratix-output
  mkdir -p $KRATIX_INPUT
  mkdir -p $KRATIX_OUTPUT

  cat <<EOF > "${KRATIX_INPUT}/object.yaml"
metadata:
  name: foo
spec:
  foo: bar
EOF

  CHART_URL=https://fluxcd-community.github.io/helm-charts CHART_NAME=flux2 $ROOT/pipeline.sh 2>&1 | grep "template foo flux2 --repo https://fluxcd-community.github.io/helm-charts --values values.yaml"
  echo "  testing helm chart from a repo with a name passed"
  rm -rf $KRATIX_INPUT
  rm -rf $KRATIX_OUTPUT
}

function cleanup {
  rm values.yaml 2> /dev/null || true
}

trap cleanup EXIT

echo "running helm promise aspect tests"
testOCI
testRepo
echo "all tests passed"