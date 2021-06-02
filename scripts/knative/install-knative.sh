#!/bin/sh

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

curl -H 'Cache-Control: no-cache' https://knative.direktiv.io/yamls/serving-crds.yaml > $dir/serving-crds.yaml
curl -H 'Cache-Control: no-cache' https://knative.direktiv.io/yamls/serving-core.yaml > $dir/serving-core.yaml
curl -H 'Cache-Control: no-cache' https://knative.direktiv.io/yamls/contour.yaml > $dir/contour.yaml
curl -H 'Cache-Control: no-cache' https://knative.direktiv.io/yamls/net-contour.yaml > $dir/net-contour.yaml

kubectl apply -f $dir/serving-crds.yaml
sleep 5
kubectl apply -f $dir/serving-core.yaml
sleep 5
kubectl apply -f $dir/contour.yaml
sleep 5
kubectl apply -f $dir/net-contour.yaml

kubectl apply -f $dir/knative-default.yaml
kubectl apply -f $dir/knative-autoscaler.yaml

kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress.class":"contour.ingress.networking.knative.dev"}}'
