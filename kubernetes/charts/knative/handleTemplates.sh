#!/bin/bash

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

curl -L -H 'Cache-Control: no-cache' https://github.com/knative/serving/releases/download/v0.25.0/serving-crds.yaml > $dir/crds/serving-crds.yaml
curl -L -H 'Cache-Control: no-cache' https://github.com/knative/serving/releases/download/v0.25.0/serving-core.yaml > $dir/templates/serving-core.yaml

# knative uses {{ }} in comments, disable them
sed -i 's/{{/{{ "{{" }}/g' $dir/templates/serving-core.yaml
sed -i 's/{{/{{ "{{" }}/g' $dir/templates/serving-core.yaml

# add helm labels
sed -i 's/^  labels:/  labels:\n    {{- include "knative.labels" . | nindent 4 }}/g' $dir/templates/serving-core.yaml

# # mark crds as helm crds
perl -0777  -pi -e  's/kind: CustomResourceDefinition\nmetadata:\n/kind: CustomResourceDefinition\nmetadata:\n  annotations:\n    helm.sh\/hook: crd-install\n/igs' $dir/templates/serving-core.yaml

# change namespace names
sed -i 's/name: knative-serving/name: knative-serving-{{ .Release.Name }}/g' $dir/templates/serving-core.yaml
sed -i 's/namespace: knative-serving/namespace: knative-serving-{{ .Release.Name }}/g' $dir/templates/serving-core.yaml

# cutting out config-autoscaler
sed -i '3577,3785d' $dir/templates/serving-core.yaml

# delete config-network
sed -i '4264,4406d' $dir/templates/serving-core.yaml
#
# # delete config-features
sed -i '3870,4018d' $dir/templates/serving-core.yaml
#
# # delete config-deployment
sed -i '3714,3803d' $dir/templates/serving-core.yaml

# add proxy settings to controller deployment env
# ed $dir/templates/serving-core.yaml < ed.script
