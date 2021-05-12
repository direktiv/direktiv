#!/bin/sh

kubectl apply -f serving-crds.yaml
sleep 5
kubectl apply -f serving-core.yaml
sleep 5
kubectl apply -f contour.yaml
sleep 5
kubectl apply -f net-contour.yaml

kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress.class":"contour.ingress.networking.knative.dev"}}'
