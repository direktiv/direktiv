#!/bin/bash

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

curl -H 'Cache-Control: no-cache' https://knative.direktiv.io/yamls/serving-crds.yaml > $dir/crds/serving-crds.yaml
curl -H 'Cache-Control: no-cache' https://knative.direktiv.io/yamls/serving-core.yaml > $dir/templates/serving-core.yaml
curl -H 'Cache-Control: no-cache' https://knative.direktiv.io/yamls/contour.yaml > $dir/templates/contour.yaml
curl -H 'Cache-Control: no-cache' https://knative.direktiv.io/yamls/net-contour.yaml > $dir/templates/net-contour.yaml

sed -i 's/^  labels:/  labels:\n    {{- include "knative.labels" . | nindent 4 }}/g' templates/contour.yaml
sed -i 's/^  labels:/  labels:\n    {{- include "knative.labels" . | nindent 4 }}/g' templates/net-contour.yaml

perl -0777  -pi -e  's/kind: CustomResourceDefinition\nmetadata:\n/kind: CustomResourceDefinition\nmetadata:\n  annotations:\n    helm.sh\/hook: crd-install\n/igs' templates/serving-core.yaml

sed -i 's/{{/{{ "{{" }}/g' templates/serving-core.yaml
sed -i 's/^  labels:/  labels:\n    {{- include "knative.labels" . | nindent 4 }}/g' templates/serving-core.yaml

sed -i 's/name: knative-serving/name: knative-serving-{{ .Release.Name }}/g' templates/serving-core.yaml
sed -i 's/namespace: knative-serving/namespace: knative-serving-{{ .Release.Name }}/g' templates/serving-core.yaml


sed -i 's/name: knative-serving/name: knative-serving-{{ .Release.Name }}/g' templates/contour.yaml

sed -i 's/namespace: knative-serving/namespace: knative-serving-{{ .Release.Name }}/g' templates/net-contour.yaml
