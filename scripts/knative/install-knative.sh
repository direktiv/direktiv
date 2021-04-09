#!/bin/sh

kubectl apply -f serving-crds.yaml
sleep 5
kubectl apply -f serving-core.yaml
sleep 5
kubectl apply -f istio.yaml
sleep 5
kubectl apply -f net-istio.yaml
