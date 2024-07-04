#!/usr/bin/env sh

set -eux

name=$(yq '.metadata.name' /kratix/input/object.yaml)
spec=$(yq '.spec' --output-format json /kratix/input/object.yaml)

cat <<EOF > request-object.yaml
apiVersion: ${OPERATOR_GROUP}/${OPERATOR_VERSION}
kind: ${OPERATOR_KIND}
metadata:
  name: ${name}
  namespace: default
spec: ${spec}
EOF

yq '.' --prettyPrint request-object.yaml > /kratix/output/object.yaml
