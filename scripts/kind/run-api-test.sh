#!/bin/bash

set -e

DEFAULT_CLUSTER_NAME="kind"
CLUSTER_NAME="${1:-$DEFAULT_CLUSTER_NAME}"

kubectl cluster-info --context kind-$CLUSTER_NAME

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
