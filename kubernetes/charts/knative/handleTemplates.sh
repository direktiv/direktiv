#!/bin/bash

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

curl -H 'Cache-Control: no-cache' https://knative.direktiv.io/yamls/serving-crds.yaml > $dir/crds/serving-crds.yaml
curl -H 'Cache-Control: no-cache' https://knative.direktiv.io/yamls/serving-core.yaml > $dir/templates/serving-core.yaml
curl -H 'Cache-Control: no-cache' https://knative.direktiv.io/yamls/contour.yaml > $dir/templates/contour.yaml
curl -H 'Cache-Control: no-cache' https://knative.direktiv.io/yamls/net-contour.yaml > $dir/templates/net-contour.yaml

# knative uses {{ }} in comments, disable them
sed -i 's/{{/{{ "{{" }}/g' $dir/templates/serving-core.yaml

# add helm labels to contour files
sed -i 's/^  labels:/  labels:\n    {{- include "knative.labels" . | nindent 4 }}/g' $dir/templates/contour.yaml
sed -i 's/^  labels:/  labels:\n    {{- include "knative.labels" . | nindent 4 }}/g' $dir/templates/net-contour.yaml
sed -i 's/^  labels:/  labels:\n    {{- include "knative.labels" . | nindent 4 }}/g' $dir/templates/serving-core.yaml

# mark crds as helm crds
perl -0777  -pi -e  's/kind: CustomResourceDefinition\nmetadata:\n/kind: CustomResourceDefinition\nmetadata:\n  annotations:\n    helm.sh\/hook: crd-install\n/igs' $dir/templates/serving-core.yaml

# change namespace names
sed -i 's/name: knative-serving/name: knative-serving-{{ .Release.Name }}/g' $dir/templates/serving-core.yaml
sed -i 's/namespace: knative-serving/namespace: knative-serving-{{ .Release.Name }}/g' $dir/templates/serving-core.yaml
sed -i 's/name: knative-serving/name: knative-serving-{{ .Release.Name }}/g' $dir/templates/contour.yaml
sed -i 's/namespace: knative-serving/namespace: knative-serving-{{ .Release.Name }}/g' $dir/templates/net-contour.yaml

# cutting out cm for autoscaler, needs to be changed per release potentially
# sed -i '3245,3436p ' $dir/templates/serving-core.yaml
sed -i '3245,3436d' $dir/templates/serving-core.yaml


ed $dir/templates/serving-core.yaml < ed.script
