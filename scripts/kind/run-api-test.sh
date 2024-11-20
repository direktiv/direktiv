#!/bin/bash

set -e

DEFAULT_CLUSTER_NAME="kind"
CLUSTER_NAME="${1:-$DEFAULT_CLUSTER_NAME}"

kubectl cluster-info --context kind-$CLUSTER_NAME

echo "Waiting for pods to be ready..."
kubectl wait --for=condition=Ready pod -l app.kubernetes.io/instance=direktiv,app.kubernetes.io/name=direktiv --namespace default --timeout=120s
kubectl wait --for=condition=Ready pod -l app.kubernetes.io/instance=direktiv,app.kubernetes.io/name=ingress-nginx --namespace default --timeout=120s
export DIREKTIV_HOST=(`kubectl get services direktiv-ingress-nginx-controller --output jsonpath='{.status.loadBalancer.ingress[0].ip}'`)
echo Direktiv is at $DIREKTIV_HOST

# Run E2E Tests
echo "Running E2E tests..."
docker run -it --network kind --rm \
  -v $(pwd)/tests:/tests \
  -e DIREKTIV_HOST=$DIREKTIV_HOST \
  -e NODE_TLS_REJECT_UNAUTHORIZED=0 \
  node:lts-alpine3.18 npm --prefix "/tests" run jest -- /tests/ --runInBand

echo "All tests completed and resources cleaned up!"
