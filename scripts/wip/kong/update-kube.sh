#!/bin/sh

kubectl apply -f scripts/wip/kong/kong-deployment.yaml
kubectl apply -f scripts/wip/kong/kong-ingress.yaml
kubectl apply -f scripts/wip/kong/kong-plugin.yaml
