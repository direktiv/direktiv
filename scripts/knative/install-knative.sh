#!/bin/sh

kubectl apply -f serving-crds.yaml
kubectl apply -f serving-core.yaml
kubectl apply -f contour.yaml
kubectl apply -f net-contour.yaml

kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress.class":"contour.ingress.networking.knative.dev"}}'
