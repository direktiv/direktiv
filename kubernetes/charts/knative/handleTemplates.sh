#!/bin/bash

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

curl -L -H 'Cache-Control: no-cache' https://github.com/knative/serving/releases/download/v0.25.0/serving-crds.yaml > $dir/crds/serving-crds.yaml
curl -L -H 'Cache-Control: no-cache' https://github.com/knative/serving/releases/download/v0.25.0/serving-core.yaml > $dir/templates/serving-core.yaml
curl -L -H 'Cache-Control: no-cache' https://github.com/knative/net-contour/releases/download/v0.25.0/contour.yaml > $dir/templates/contour.yaml
curl -L -H 'Cache-Control: no-cache' https://github.com/knative/net-contour/releases/download/v0.25.0/net-contour.yaml > $dir/templates/net-contour.yaml

# knative uses {{ }} in comments, disable them
sed -i 's/{{/{{ "{{" }}/g' $dir/templates/serving-core.yaml
sed -i 's/{{/{{ "{{" }}/g' $dir/templates/serving-core.yaml

# add helm labels to contour files
sed -i 's/prometheus.io\/scrape: "true"/prometheus.io\/scrape: "false"/g' $dir/templates/contour.yaml
sed -i 's/^  labels:/  labels:\n    {{- include "knative.labels" . | nindent 4 }}/g' $dir/templates/contour.yaml
sed -i 's/^  labels:/  labels:\n    {{- include "knative.labels" . | nindent 4 }}/g' $dir/templates/net-contour.yaml
sed -i 's/^  labels:/  labels:\n    {{- include "knative.labels" . | nindent 4 }}/g' $dir/templates/serving-core.yaml

# # mark crds as helm crds
perl -0777  -pi -e  's/kind: CustomResourceDefinition\nmetadata:\n/kind: CustomResourceDefinition\nmetadata:\n  annotations:\n    helm.sh\/hook: crd-install\n/igs' $dir/templates/serving-core.yaml

# change namespace names
sed -i 's/name: knative-serving/name: knative-serving-{{ .Release.Name }}/g' $dir/templates/serving-core.yaml
sed -i 's/namespace: knative-serving/namespace: knative-serving-{{ .Release.Name }}/g' $dir/templates/serving-core.yaml
sed -i 's/name: knative-serving/name: knative-serving-{{ .Release.Name }}/g' $dir/templates/contour.yaml
sed -i 's/namespace: knative-serving/namespace: knative-serving-{{ .Release.Name }}/g' $dir/templates/net-contour.yaml

# cutting out config-autoscaler
sed -i '3576,3784d' $dir/templates/serving-core.yaml

# delete config-network
sed -i '4263,4405d' $dir/templates/serving-core.yaml

# delete config-features
sed -i '3869,4011d' $dir/templates/serving-core.yaml

# delete config-deployment
sed -i '3713,3802d' $dir/templates/serving-core.yaml

ed $dir/templates/serving-core.yaml < ed.script

# replace replicas
sed -i 's/replicas: 2/replicas: 1/g' $dir/templates/contour.yaml
