#!/bin/sh

curl https://knative.direktiv.io/yamls/serving-crds.yaml > serving-crds.yaml
curl https://knative.direktiv.io/yamls/serving-core.yaml > serving-core.yaml
curl https://knative.direktiv.io/yamls/contour.yaml > contour.yaml
curl https://knative.direktiv.io/yamls/net-contour.yaml > net-contour.yaml

kubectl apply -f serving-crds.yaml
sleep 5
kubectl apply -f serving-core.yaml
sleep 5
kubectl apply -f contour.yaml
sleep 5
kubectl apply -f net-contour.yaml

kubectl apply -f knative-default.yaml

kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress.class":"contour.ingress.networking.knative.dev"}}'
